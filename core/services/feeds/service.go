package feeds

import (
	"context"

	"github.com/pkg/errors"
	"github.com/smartcontractkit/chainlink/core/logger"
	pb "github.com/smartcontractkit/chainlink/core/services/feeds/proto"
	"github.com/smartcontractkit/chainlink/core/services/fluxmonitorv2"
	"github.com/smartcontractkit/chainlink/core/services/job"
	"github.com/smartcontractkit/chainlink/core/services/keystore"
	"github.com/smartcontractkit/chainlink/core/services/offchainreporting"
	"github.com/smartcontractkit/chainlink/core/services/postgres"
	"github.com/smartcontractkit/chainlink/core/services/versioning"
	"github.com/smartcontractkit/chainlink/core/utils"
)

//go:generate mockery --name Service --output ./mocks/ --case=underscore
//go:generate mockery --dir ./proto --name FeedsManagerClient --output ./mocks/ --case=underscore

var (
	ErrOCRDisabled = errors.New("ocr is disabled")
)

type Service interface {
	Start() error
	Close() error

	ApproveJobProposal(ctx context.Context, id int64) error
	CountManagers() (int64, error)
	CreateJobProposal(jp *JobProposal) (int64, error)
	GetJobProposal(id int64) (*JobProposal, error)
	GetManager(id int64) (*FeedsManager, error)
	ListManagers() ([]FeedsManager, error)
	ListJobProposals() ([]JobProposal, error)
	RegisterManager(ms *FeedsManager) (int64, error)
	RejectJobProposal(ctx context.Context, id int64) error
	SyncNodeInfo(id int64) error
	UpdateJobProposalSpec(ctx context.Context, id int64, spec string) error
	UpdateFeedsManager(ctx context.Context, mgr FeedsManager) error

	Unsafe_SetConnectionsManager(ConnectionsManager)
}

type service struct {
	utils.StartStopOnce

	orm         ORM
	verORM      versioning.ORM
	csaKeyStore keystore.CSAKeystoreInterface
	ethKeyStore keystore.EthKeyStoreInterface
	jobSpawner  job.Spawner
	cfg         Config
	txm         postgres.TransactionManager
	connMgr     ConnectionsManager
}

// NewService constructs a new feeds service
func NewService(
	orm ORM,
	verORM versioning.ORM,
	txm postgres.TransactionManager,
	jobSpawner job.Spawner,
	csaKeyStore keystore.CSAKeystoreInterface,
	ethKeyStore keystore.EthKeyStoreInterface,
	cfg Config,
) *service {
	svc := &service{
		orm:         orm,
		verORM:      verORM,
		txm:         txm,
		jobSpawner:  jobSpawner,
		csaKeyStore: csaKeyStore,
		ethKeyStore: ethKeyStore,
		cfg:         cfg,
		connMgr:     newConnectionsManager(),
	}

	return svc
}

// RegisterManager registers a new ManagerService and attempts to establish a
// connection.
//
// Only a single feeds manager is currently supported.
func (s *service) RegisterManager(mgr *FeedsManager) (int64, error) {
	count, err := s.CountManagers()
	if err != nil {
		return 0, err
	}
	if count >= 1 {
		return 0, errors.New("only a single feeds manager is supported")
	}

	id, err := s.orm.CreateManager(context.Background(), mgr)
	if err != nil {
		return 0, err
	}

	privkey, err := s.getCSAPrivateKey()
	if err != nil {
		return 0, err
	}

	// Establish a connection
	mgr.ID = id
	s.connectFeedManager(*mgr, privkey)

	return id, nil
}

// SyncNodeInfo syncs the node's information with FMS
func (s *service) SyncNodeInfo(id int64) error {
	mgr, err := s.GetManager(id)
	if err != nil {
		return err
	}

	jobtypes := []pb.JobType{}
	for _, jt := range mgr.JobTypes {
		switch jt {
		case JobTypeFluxMonitor:
			jobtypes = append(jobtypes, pb.JobType_JOB_TYPE_FLUX_MONITOR)
		case JobTypeOffchainReporting:
			jobtypes = append(jobtypes, pb.JobType_JOB_TYPE_OCR)
		default:
			// NOOP
		}
	}

	keys, err := s.ethKeyStore.SendingKeys()
	if err != nil {
		return err
	}

	addresses := []string{}
	for _, k := range keys {
		addresses = append(addresses, k.Address.String())
	}

	nodeVer, err := s.verORM.FindLatestNodeVersion()
	if err != nil {
		return errors.Wrap(err, "could not get latest node verion")
	}

	// Make the remote call to FMS
	fmsClient, err := s.connMgr.GetClient(id)
	if err != nil {
		return errors.Wrap(err, "could not fetch client")
	}

	_, err = fmsClient.UpdateNode(context.Background(), &pb.UpdateNodeRequest{
		JobTypes:           jobtypes,
		ChainId:            s.cfg.ChainID().Int64(),
		AccountAddresses:   addresses,
		IsBootstrapPeer:    mgr.IsOCRBootstrapPeer,
		BootstrapMultiaddr: mgr.OCRBootstrapPeerMultiaddr.ValueOrZero(),
		Version:            nodeVer.Version,
	})
	if err != nil {
		return err
	}

	return nil
}

// UpdateFeedsManager updates the feed manager details, takes down the
// connection and reestablishes a new connection with the updated public key.
func (s *service) UpdateFeedsManager(ctx context.Context, mgr FeedsManager) error {
	ctx, cancel := context.WithTimeout(ctx, postgres.DefaultQueryTimeout)
	defer cancel()

	err := s.txm.TransactWithContext(ctx, func(ctx context.Context) error {
		err := s.orm.UpdateManager(ctx, mgr)
		if err != nil {
			return errors.Wrap(err, "could not update manager")
		}

		return nil
	})
	if err != nil {
		return err
	}

	if err = s.connMgr.Disconnect(mgr.ID); err != nil {
		logger.Info("[Feeds] Feeds Manager not connected, attempting to connect")
	}

	// Establish a new connection
	privkey, err := s.getCSAPrivateKey()
	if err != nil {
		return err
	}

	s.connectFeedManager(mgr, privkey)

	return nil
}

// ListManagerServices lists all the manager services.
func (s *service) ListManagers() ([]FeedsManager, error) {
	managers, err := s.orm.ListManagers(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "failed to get a list of managers")
	}

	for i := range managers {
		managers[i].IsConnectionActive = s.connMgr.IsConnected(managers[i].ID)
	}

	return managers, nil
}

// GetManager gets a manager service by id.
func (s *service) GetManager(id int64) (*FeedsManager, error) {
	manager, err := s.orm.GetManager(context.Background(), id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get manager by ID")
	}

	manager.IsConnectionActive = s.connMgr.IsConnected(manager.ID)
	return manager, nil
}

// CountManagerServices gets the total number of manager services
func (s *service) CountManagers() (int64, error) {
	return s.orm.CountManagers(context.Background())
}

// Lists all JobProposals
//
// When we support multiple feed managers, we will need to change this to filter
// by feeds manager
func (s *service) ListJobProposals() ([]JobProposal, error) {
	return s.orm.ListJobProposals(context.Background())
}

// CreateJobProposal creates a job proposal.
func (s *service) CreateJobProposal(jp *JobProposal) (int64, error) {
	// Bootstrap multiaddrs only allowed for OCR jobs
	if len(jp.Multiaddrs) > 0 {
		j, err := s.generateJob(jp.Spec)
		if err != nil {
			return 0, errors.Wrap(err, "failed to generate a job based on spec")
		}

		if j.Type != job.OffchainReporting {
			return 0, errors.New("only OCR job type supports multiaddr")
		}
	}

	return s.orm.CreateJobProposal(context.Background(), jp)
}

// GetJobProposal gets a job proposal by id.
func (s *service) GetJobProposal(id int64) (*JobProposal, error) {
	return s.orm.GetJobProposal(context.Background(), id)
}

func (s *service) UpdateJobProposalSpec(ctx context.Context, id int64, spec string) error {
	jp, err := s.orm.GetJobProposal(ctx, id)
	if err != nil {
		return errors.Wrap(err, "job proposal does not exist")
	}

	if jp.Status != JobProposalStatusPending {
		return errors.New("must be a pending job proposal")
	}

	// Update the spec
	if err = s.orm.UpdateJobProposalSpec(ctx, id, spec); err != nil {
		return errors.Wrap(err, "could not update job proposal")
	}

	return nil
}

func (s *service) ApproveJobProposal(ctx context.Context, id int64) error {
	jp, err := s.orm.GetJobProposal(ctx, id)
	if err != nil {
		return errors.Wrap(err, "job proposal does not exist")
	}

	fmsClient, err := s.connMgr.GetClient(jp.FeedsManagerID)
	if err != nil {
		return errors.Wrap(err, "fms rpc client is not connected")
	}

	if jp.Status != JobProposalStatusPending {
		return errors.New("must be a pending job proposal")
	}

	ctx, cancel := context.WithTimeout(ctx, postgres.DefaultQueryTimeout)
	defer cancel()

	j, err := s.generateJob(jp.Spec)
	if err != nil {
		return errors.Wrap(err, "could not generate job from spec")
	}

	err = s.txm.TransactWithContext(ctx, func(ctx context.Context) error {
		// Create the job
		_, err = s.jobSpawner.CreateJob(ctx, *j, j.Name)
		if err != nil {
			return err
		}

		// Approve the job
		if err = s.orm.ApproveJobProposal(ctx, id, j.ExternalJobID, JobProposalStatusApproved); err != nil {
			return err
		}

		// Send to FMS Client
		if _, err = fmsClient.ApprovedJob(ctx, &pb.ApprovedJobRequest{
			Uuid: jp.RemoteUUID.String(),
		}); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return errors.Wrap(err, "could not approve job proposal")
	}

	return nil
}

func (s *service) RejectJobProposal(ctx context.Context, id int64) error {
	jp, err := s.orm.GetJobProposal(ctx, id)
	if err != nil {
		return errors.Wrap(err, "job proposal does not exist")
	}

	fmsClient, err := s.connMgr.GetClient(jp.FeedsManagerID)
	if err != nil {
		return errors.Wrap(err, "fms rpc client is not connected")
	}

	if jp.Status != JobProposalStatusPending {
		return errors.New("must be a pending job proposal")
	}

	ctx, cancel := context.WithTimeout(ctx, postgres.DefaultQueryTimeout)
	defer cancel()
	err = s.txm.TransactWithContext(ctx, func(ctx context.Context) error {
		if err = s.orm.UpdateJobProposalStatus(ctx, id, JobProposalStatusRejected); err != nil {
			return err
		}

		if _, err = fmsClient.RejectedJob(ctx, &pb.RejectedJobRequest{
			Uuid: jp.RemoteUUID.String(),
		}); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return errors.Wrap(err, "could not reject job proposal")
	}

	return nil
}

func (s *service) Start() error {
	return s.StartOnce("FeedsService", func() error {
		privkey, err := s.getCSAPrivateKey()
		if err != nil {
			return err
		}

		// We only support a single feeds manager right now
		mgrs, err := s.ListManagers()
		if err != nil {
			return err
		}
		if len(mgrs) < 1 {
			return errors.New("no feeds managers registered")
		}

		mgr := mgrs[0]
		s.connectFeedManager(mgr, privkey)

		return nil
	})
}

func (s *service) Close() error {
	return s.StopOnce("FeedsService", func() error {
		// This blocks until it finishes
		s.connMgr.Close()

		return nil
	})
}

// connectFeedManager connects to a feeds manager
func (s *service) connectFeedManager(mgr FeedsManager, privkey []byte) {
	s.connMgr.Connect(ConnectOpts{
		FeedsManagerID: mgr.ID,
		URI:            mgr.URI,
		Privkey:        privkey,
		Pubkey:         mgr.PublicKey,
		Handlers: &RPCHandlers{
			feedsManagerID: mgr.ID,
			svc:            s,
		},
		OnConnect: func(pb.FeedsManagerClient) {
			// Sync the node's information with FMS once connected
			err := s.SyncNodeInfo(mgr.ID)
			if err != nil {
				logger.Infof("[Feeds] Error syncing node info: %v", err)
			}
		},
	})
}

// getCSAPrivateKey gets the server's CSA private key
func (s *service) getCSAPrivateKey() (privkey []byte, err error) {
	// Fetch the server's public key
	keys, err := s.csaKeyStore.ListCSAKeys()
	if err != nil {
		return privkey, err
	}
	if len(keys) < 1 {
		return privkey, errors.New("CSA key does not exist")
	}

	privkey, err = s.csaKeyStore.Unsafe_GetUnlockedPrivateKey(keys[0].PublicKey)
	if err != nil {
		return []byte{}, err
	}

	return privkey, nil
}

// Unsafe_SetFMSClient sets the FMSClient on the service.
//
// We need to be able to inject a mock for the client to facilitate integration
// tests.
//
// ONLY TO BE USED FOR TESTING.
func (s *service) Unsafe_SetConnectionsManager(connMgr ConnectionsManager) {
	s.connMgr = connMgr
}

func (s *service) generateJob(spec string) (*job.Job, error) {
	jobType, err := job.ValidateSpec(spec)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse job spec TOML")
	}

	var js job.Job
	switch jobType {
	case job.OffchainReporting:
		js, err = offchainreporting.ValidatedOracleSpecToml(s.cfg, spec)
		if !s.cfg.Dev() && !s.cfg.FeatureOffchainReporting() {
			return nil, ErrOCRDisabled
		}
	case job.FluxMonitor:
		js, err = fluxmonitorv2.ValidatedFluxMonitorSpec(s.cfg, spec)
	default:
		return nil, errors.Errorf("unknown job type: %s", jobType)

	}
	if err != nil {
		return nil, err
	}

	return &js, nil
}
