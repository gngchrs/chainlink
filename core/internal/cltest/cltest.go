package cltest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/gin-gonic/gin"
	"github.com/gobuffalo/packr"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	cryptop2p "github.com/libp2p/go-libp2p-core/crypto"
	p2ppeer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/manyminds/api2go/jsonapi"
	"github.com/onsi/gomega"
	uuid "github.com/satori/go.uuid"
	"github.com/smartcontractkit/chainlink/core/auth"
	"github.com/smartcontractkit/chainlink/core/cmd"
	"github.com/smartcontractkit/chainlink/core/gracefulpanic"
	"github.com/smartcontractkit/chainlink/core/internal/gethwrappers/generated/flux_aggregator_wrapper"
	"github.com/smartcontractkit/chainlink/core/internal/mocks"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/chainlink/core/services/bulletprooftxmanager"
	"github.com/smartcontractkit/chainlink/core/services/chainlink"
	"github.com/smartcontractkit/chainlink/core/services/eth"
	"github.com/smartcontractkit/chainlink/core/services/gas"
	httypes "github.com/smartcontractkit/chainlink/core/services/headtracker/types"
	"github.com/smartcontractkit/chainlink/core/services/job"
	"github.com/smartcontractkit/chainlink/core/services/keystore"
	"github.com/smartcontractkit/chainlink/core/services/keystore/keys/ethkey"
	"github.com/smartcontractkit/chainlink/core/services/keystore/keys/p2pkey"
	"github.com/smartcontractkit/chainlink/core/services/pipeline"
	"github.com/smartcontractkit/chainlink/core/services/postgres"
	"github.com/smartcontractkit/chainlink/core/services/webhook"
	"github.com/smartcontractkit/chainlink/core/static"
	strpkg "github.com/smartcontractkit/chainlink/core/store"
	"github.com/smartcontractkit/chainlink/core/store/config"
	"github.com/smartcontractkit/chainlink/core/store/dialects"
	"github.com/smartcontractkit/chainlink/core/store/models"
	"github.com/smartcontractkit/chainlink/core/utils"
	"github.com/smartcontractkit/chainlink/core/web"
	webpresenters "github.com/smartcontractkit/chainlink/core/web/presenters"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"go.uber.org/zap/zapcore"
	null "gopkg.in/guregu/null.v4"
	"gorm.io/gorm"

	"github.com/smartcontractkit/chainlink/core/internal/testutils/configtest"
	// Force import of pgtest to ensure that txdb is registered as a DB driver
	_ "github.com/smartcontractkit/chainlink/core/internal/testutils/pgtest"
)

const (
	// APIKey of the fixture API user
	APIKey = "2d25e62eaf9143e993acaf48691564b2"
	// APISecret of the fixture API user.
	APISecret = "1eCP/w0llVkchejFaoBpfIGaLRxZK54lTXBCT22YLW+pdzE4Fafy/XO5LoJ2uwHi"
	// APIEmail is the email of the fixture API user
	APIEmail = "apiuser@chainlink.test"
	// Password just a password we use everywhere for testing
	Password    = "p4SsW0rD1!@#_"
	VRFPassword = "testingpassword"
	// SessionSecret is the hardcoded secret solely used for test
	SessionSecret = "clsession_test_secret"
	// DefaultKeyAddress is the ETH address of the fixture key
	DefaultKeyAddress = "0xF67D0290337bca0847005C7ffD1BC75BA9AAE6e4"
	// DefaultKeyFixtureFileName is the filename of the fixture key
	DefaultKeyFixtureFileName = "testkey-0xF67D0290337bca0847005C7ffD1BC75BA9AAE6e4.json"
	// DefaultKeyJSON is the JSON for the default key encrypted with fast scrypt and password 'password' (used for fixture file)
	DefaultKeyJSON = `{"address":"F67D0290337bca0847005C7ffD1BC75BA9AAE6e4","crypto":{"cipher":"aes-128-ctr","ciphertext":"9c3565050ba4e10ea388bcd17d77c141441ce1be5db339f0201b9ed733d780c6","cipherparams":{"iv":"f968fc947495646ee8b5dbaadb242ec0"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"33ad88742a983dfeb8adcc9a39fdde4cb47f7e23ea2ef80b35723d940959e3fd"},"mac":"b3747959cbbb9b26f861ab82d69154b4ec8108bbac017c1341f6fd3295beceaf"},"id":"8c79a654-96b1-45d9-8978-3efa07578011","version":3}`
	// AllowUnstarted enable an application that can be used in tests without being started
	AllowUnstarted = "allow_unstarted"
	// A peer ID without an associated p2p key.
	NonExistentPeerID = "12D3KooWAdCzaesXyezatDzgGvCngqsBqoUqnV9PnVc46jsVt2i9"
	// DefaultOCRKeyBundleID is the ID of the fixture ocr key bundle
	DefaultOCRKeyBundleID = "7f993fb701b3410b1f6e8d4d93a7462754d24609b9b31a4fe64a0cb475a4d934"
)

var (
	DefaultP2PPeerID     p2pkey.PeerID
	NonExistentP2PPeerID p2pkey.PeerID
	// DefaultOCRKeyBundleIDSha256 is the ID of the fixture ocr key bundle
	DefaultOCRKeyBundleIDSha256 models.Sha256Hash
	FluxAggAddress              = common.HexToAddress("0x3cCad4715152693fE3BC4460591e3D3Fbd071b42")
	source                      rand.Source
)

func init() {
	gin.SetMode(gin.TestMode)
	gomega.SetDefaultEventuallyTimeout(3 * time.Second)
	lvl := logLevelFromEnv()
	logger.SetLogger(logger.CreateTestLogger(lvl))

	// Seed the random number generator, otherwise separate modules will take
	// the same advisory locks when tested with `go test -p N` for N > 1
	seed := time.Now().UTC().UnixNano()
	logger.Debugf("Using seed: %v", seed)
	rand.Seed(seed)

	// Also seed the local source
	source = rand.NewSource(seed)

	defaultP2PPeerID, err := p2ppeer.Decode(configtest.DefaultPeerID)
	if err != nil {
		panic(err)
	}
	DefaultP2PPeerID = p2pkey.PeerID(defaultP2PPeerID)
	nonExistentP2PPeerID, err := p2ppeer.Decode(NonExistentPeerID)
	if err != nil {
		panic(err)
	}
	NonExistentP2PPeerID = p2pkey.PeerID(nonExistentP2PPeerID)
	DefaultOCRKeyBundleIDSha256, err = models.Sha256HashFromHex(DefaultOCRKeyBundleID)
	if err != nil {
		panic(err)
	}
}

func logLevelFromEnv() zapcore.Level {
	lvl := zapcore.ErrorLevel
	if env := os.Getenv("LOG_LEVEL"); env != "" {
		_ = lvl.Set(env)
	}
	return lvl
}

func NewRandomInt64() int64 {
	id := rand.Int63()
	return id
}

func MustRandomBytes(t *testing.T, l int) (b []byte) {
	t.Helper()

	b = make([]byte, l)
	/* #nosec G404 */
	_, err := rand.Read(b)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func MustBigIntFromString(t *testing.T, s string) *big.Int {
	t.Helper()
	x, ok := big.NewInt(0).SetString(s, 10)
	if !ok {
		t.Fatalf("could not create *big.Int from string '%v'", s)
	}
	return x
}

type JobPipelineV2TestHelper struct {
	Prm pipeline.ORM
	Eb  postgres.EventBroadcaster
	Jrm job.ORM
	Pr  pipeline.Runner
}

func NewJobPipelineV2(t testing.TB, cfg config.EVMConfig, db *gorm.DB, ethClient eth.Client, keyStore pipeline.ETHKeyStore, txManager pipeline.TxManager) JobPipelineV2TestHelper {
	prm, eb, cleanup := NewPipelineORM(t, cfg, db)
	jrm := job.NewORM(db, cfg, prm, eb, &postgres.NullAdvisoryLocker{})
	t.Cleanup(cleanup)
	pr := pipeline.NewRunner(prm, cfg, ethClient, keyStore, nil, txManager)
	return JobPipelineV2TestHelper{
		prm,
		eb,
		jrm,
		pr,
	}
}

func NewPipelineORM(t testing.TB, cfg config.GeneralConfig, db *gorm.DB) (pipeline.ORM, postgres.EventBroadcaster, func()) {
	t.Helper()
	eventBroadcaster := postgres.NewEventBroadcaster(cfg.DatabaseURL(), 0, 0)
	err := eventBroadcaster.Start()
	require.NoError(t, err)
	return pipeline.NewORM(db), eventBroadcaster, func() {
		eventBroadcaster.Close()
	}
}

func NewEthBroadcaster(t testing.TB, db *gorm.DB, ethClient eth.Client, keyStore bulletprooftxmanager.KeyStore, config config.EVMConfig, keys ...ethkey.Key) (*bulletprooftxmanager.EthBroadcaster, func()) {
	t.Helper()
	eventBroadcaster := postgres.NewEventBroadcaster(config.DatabaseURL(), 0, 0)
	err := eventBroadcaster.Start()
	require.NoError(t, err)
	return bulletprooftxmanager.NewEthBroadcaster(db, ethClient, config, keyStore, &postgres.NullAdvisoryLocker{}, eventBroadcaster, keys, gas.NewFixedPriceEstimator(config)), func() {
		assert.NoError(t, eventBroadcaster.Close())
	}
}

func NewEthConfirmer(t testing.TB, db *gorm.DB, ethClient eth.Client, config config.EVMConfig, ks bulletprooftxmanager.KeyStore, keys []ethkey.Key) *bulletprooftxmanager.EthConfirmer {
	t.Helper()
	ec := bulletprooftxmanager.NewEthConfirmer(db, ethClient, config, ks, &postgres.NullAdvisoryLocker{}, keys, gas.NewFixedPriceEstimator(config))
	return ec
}

// TestApplication holds the test application and test servers
type TestApplication struct {
	t testing.TB
	*chainlink.ChainlinkApplication
	Logger         *logger.Logger
	Server         *httptest.Server
	wsServer       *httptest.Server
	Started        bool
	Backend        *backends.SimulatedBackend
	Key            ethkey.Key
	allowUnstarted bool
}

// NewWSServer returns a  new wsserver
func NewWSServer(msg string, callback func(data []byte)) (*httptest.Server, string, func()) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		logger.PanicIf(err)
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				break
			}

			if callback != nil {
				callback(data)
			}

			err = conn.WriteMessage(websocket.BinaryMessage, []byte(msg))
			if err != nil {
				break
			}
		}
	})
	server := httptest.NewServer(handler)

	u, err := url.Parse(server.URL)
	logger.PanicIf(err)
	u.Scheme = "ws"

	return server, u.String(), func() {
		server.Close()
	}
}

func NewTestEVMConfig(t testing.TB) *configtest.TestEVMConfig {
	overrides := configtest.GeneralConfigOverrides{
		SecretGenerator: MockSecretGenerator{},
		Dialect:         dialects.TransactionWrappedPostgres,
		AdvisoryLockID:  null.IntFrom(NewRandomInt64()),
	}
	cfg := configtest.NewTestGeneralConfigWithOverrides(t, overrides)
	evmcfg := configtest.NewTestEVMConfig(t, cfg)
	return evmcfg
}

// NewApplicationEthereumDisabled creates a new application with default config but ethereum disabled
// Useful for testing controllers
func NewApplicationEthereumDisabled(t *testing.T) (*TestApplication, func()) {
	t.Helper()

	c := NewTestEVMConfig(t)
	c.GeneralConfig.Overrides.EthereumDisabled = null.BoolFrom(true)

	app, cleanup := NewApplicationWithConfig(t, c)

	return app, cleanup
}

// NewApplication creates a New TestApplication along with a NewConfig
// It mocks the keystore with no keys or accounts by default
func NewApplication(t testing.TB, flagsAndDeps ...interface{}) (*TestApplication, func()) {
	t.Helper()

	c := NewTestEVMConfig(t)

	app, cleanup := NewApplicationWithConfig(t, c, flagsAndDeps...)

	return app, cleanup
}

// NewApplicationWithKey creates a new TestApplication along with a new config
// It uses the native keystore and will load any keys that are in the database
func NewApplicationWithKey(t *testing.T, flagsAndDeps ...interface{}) (*TestApplication, func()) {
	t.Helper()

	config := NewTestEVMConfig(t)
	app, cleanup := NewApplicationWithConfigAndKey(t, config, flagsAndDeps...)
	return app, cleanup
}

// NewApplicationWithConfigAndKey creates a new TestApplication with the given testorm
// it will also provide an unlocked account on the keystore
func NewApplicationWithConfigAndKey(t testing.TB, c *configtest.TestEVMConfig, flagsAndDeps ...interface{}) (*TestApplication, func()) {
	t.Helper()

	app, cleanup := NewApplicationWithConfig(t, c, flagsAndDeps...)
	for _, dep := range flagsAndDeps {
		switch v := dep.(type) {
		case ethkey.Key:
			MustAddKeyToKeystore(t, &v, app.KeyStore.Eth())
			app.Key = v
		}
	}
	if app.Key.Address.Address() == utils.ZeroAddress {
		app.Key, _ = MustAddRandomKeyToKeystore(t, app.KeyStore.Eth(), 0)
	}
	require.NoError(t, app.KeyStore.Eth().Unlock(Password))
	_, err := app.KeyStore.VRF().Unlock(VRFPassword)
	require.NoError(t, err)

	return app, cleanup
}

const (
	UseRealExternalInitiatorManager = "UseRealExternalInitiatorManager"
)

// NewApplicationWithConfig creates a New TestApplication with specified test config
func NewApplicationWithConfig(t testing.TB, c *configtest.TestEVMConfig, flagsAndDeps ...interface{}) (*TestApplication, func()) {
	t.Helper()

	var ethClient eth.Client = &eth.NullClient{}
	var advisoryLocker postgres.AdvisoryLocker = &postgres.NullAdvisoryLocker{}
	var externalInitiatorManager webhook.ExternalInitiatorManager = &webhook.NullExternalInitiatorManager{}
	var useRealExternalInitiatorManager bool

	for _, flag := range flagsAndDeps {
		switch dep := flag.(type) {
		case eth.Client:
			ethClient = dep
		case postgres.AdvisoryLocker:
			advisoryLocker = dep
		case webhook.ExternalInitiatorManager:
			externalInitiatorManager = dep
		default:
			switch flag {
			case UseRealExternalInitiatorManager:
				useRealExternalInitiatorManager = true
			}

		}
	}

	ta := &TestApplication{t: t}
	appInstance, err := chainlink.NewApplication(c, ethClient, advisoryLocker)
	require.NoError(t, err)
	app := appInstance.(*chainlink.ChainlinkApplication)
	ta.ChainlinkApplication = app
	server := newServer(ta)

	c.GeneralConfig.Overrides.ClientNodeURL = null.StringFrom(server.URL)

	if !useRealExternalInitiatorManager {
		app.ExternalInitiatorManager = externalInitiatorManager
	}

	for _, flag := range flagsAndDeps {
		if flag == AllowUnstarted {
			ta.allowUnstarted = true
		}
	}

	ta.Server = server
	return ta, func() {
		err := ta.StopIfStarted()
		require.NoError(t, err)
	}
}

func NewEthMocks(t testing.TB) (*mocks.Client, *mocks.Subscription, func()) {
	c, s := NewEthClientAndSubMock(t)
	var assertMocksCalled func()
	switch tt := t.(type) {
	case *testing.T:
		assertMocksCalled = func() {
			c.AssertExpectations(tt)
			s.AssertExpectations(tt)
		}
	case *testing.B:
		assertMocksCalled = func() {}
	}
	return c, s, assertMocksCalled
}

func NewEthClientAndSubMock(t mock.TestingT) (*mocks.Client, *mocks.Subscription) {
	mockSub := new(mocks.Subscription)
	mockSub.Test(t)
	mockEth := new(mocks.Client)
	mockEth.Test(t)
	return mockEth, mockSub
}

func NewEthClientMock(t mock.TestingT) *mocks.Client {
	mockEth := new(mocks.Client)
	mockEth.Test(t)
	return mockEth
}

func NewEthMocksWithStartupAssertions(t testing.TB) (*mocks.Client, *mocks.Subscription, func()) {
	c, s, assertMocksCalled := NewEthMocks(t)
	c.On("Dial", mock.Anything).Maybe().Return(nil)
	c.On("SubscribeNewHead", mock.Anything, mock.Anything).Maybe().Return(EmptyMockSubscription(), nil)
	c.On("SendTransaction", mock.Anything, mock.Anything).Maybe().Return(nil)
	c.On("HeadByNumber", mock.Anything, (*big.Int)(nil)).Maybe().Return(Head(0), nil)
	c.On("ChainID", mock.Anything).Maybe().Return(big.NewInt(eth.NullClientChainID), nil)

	block := types.NewBlockWithHeader(&types.Header{
		Number: big.NewInt(100),
	})
	c.On("BlockByNumber", mock.Anything, mock.Anything).Maybe().Return(block, nil)

	s.On("Err").Return(nil).Maybe()
	s.On("Unsubscribe").Return(nil).Maybe()
	return c, s, assertMocksCalled
}

func newServer(app chainlink.Application) *httptest.Server {
	engine := web.Router(app)
	return httptest.NewServer(engine)
}

func (ta *TestApplication) NewBox() packr.Box {
	ta.t.Helper()

	return packr.NewBox("../fixtures/operator_ui/dist")
}

func (ta *TestApplication) Start() error {
	ta.t.Helper()
	ta.Started = true
	// TODO - RYAN - we should have a global keystore.Unlock() function
	// https://app.clubhouse.io/chainlinklabs/story/7735/combine-keystores
	err := ta.ChainlinkApplication.KeyStore.Eth().Unlock(Password)
	if err != nil {
		return err
	}

	err = ta.ChainlinkApplication.Start()
	return err
}

// Stop will stop the test application and perform cleanup
func (ta *TestApplication) Stop() error {
	ta.t.Helper()

	if !ta.Started {
		if ta.allowUnstarted {
			return nil
		}
		ta.t.Fatal("TestApplication Stop() called on an unstarted application")
	}

	// TODO: Here we double close, which is less than ideal.
	// We would prefer to invoke a method on an interface that
	// cleans up only in test.
	err := ta.ChainlinkApplication.StopIfStarted()
	if err != nil {
		return err
	}
	cleanUpStore(ta.t, ta.Store)
	if ta.Server != nil {
		ta.Server.Close()
	}
	if ta.wsServer != nil {
		ta.wsServer.Close()
	}
	return nil
}

func (ta *TestApplication) MustSeedNewSession() string {
	session := NewSession()
	require.NoError(ta.t, ta.Store.DB.Save(&session).Error)
	return session.ID
}

// ImportKey adds private key to the application disk keystore, not database.
func (ta *TestApplication) ImportKey(content string) {
	_, err := ta.KeyStore.Eth().ImportKey([]byte(content), Password)
	require.NoError(ta.t, err)
	require.NoError(ta.t, ta.KeyStore.Eth().Unlock(Password))
}

func (ta *TestApplication) NewHTTPClient() HTTPClientCleaner {
	ta.t.Helper()

	sessionID := ta.MustSeedNewSession()

	return HTTPClientCleaner{
		HTTPClient: NewMockAuthenticatedHTTPClient(ta.Config, sessionID),
		t:          ta.t,
	}
}

// NewClientAndRenderer creates a new cmd.Client for the test application
func (ta *TestApplication) NewClientAndRenderer() (*cmd.Client, *RendererMock) {
	sessionID := ta.MustSeedNewSession()
	r := &RendererMock{}
	client := &cmd.Client{
		Renderer:   r,
		Config:     ta.GetEVMConfig(),
		AppFactory: seededAppFactory{ta.ChainlinkApplication},
		KeyStoreAuthenticator: CallbackAuthenticator{
			Callback: func(*keystore.Eth, string) (string, error) {
				return Password, nil
			},
		},
		FallbackAPIInitializer:         &MockAPIInitializer{},
		Runner:                         EmptyRunner{},
		HTTP:                           NewMockAuthenticatedHTTPClient(ta.Config, sessionID),
		CookieAuthenticator:            MockCookieAuthenticator{},
		FileSessionRequestBuilder:      &MockSessionRequestBuilder{},
		PromptingSessionRequestBuilder: &MockSessionRequestBuilder{},
		ChangePasswordPrompter:         &MockChangePasswordPrompter{},
	}
	return client, r
}

func (ta *TestApplication) NewAuthenticatingClient(prompter cmd.Prompter) *cmd.Client {
	cookieAuth := cmd.NewSessionCookieAuthenticator(ta.GetConfig(), &cmd.MemoryCookieStore{})
	client := &cmd.Client{
		Renderer:                       &RendererMock{},
		Config:                         ta.GetEVMConfig(),
		AppFactory:                     seededAppFactory{ta.ChainlinkApplication},
		KeyStoreAuthenticator:          CallbackAuthenticator{func(*keystore.Eth, string) (string, error) { return Password, nil }},
		FallbackAPIInitializer:         &MockAPIInitializer{},
		Runner:                         EmptyRunner{},
		HTTP:                           cmd.NewAuthenticatedHTTPClient(ta.Config, cookieAuth, models.SessionRequest{}),
		CookieAuthenticator:            cookieAuth,
		FileSessionRequestBuilder:      cmd.NewFileSessionRequestBuilder(),
		PromptingSessionRequestBuilder: cmd.NewPromptingSessionRequestBuilder(prompter),
		ChangePasswordPrompter:         &MockChangePasswordPrompter{},
	}
	return client
}

// NewStoreWithConfig creates a new store with given config
func NewStoreWithConfig(t testing.TB, c config.GeneralConfig, flagsAndDeps ...interface{}) (*strpkg.Store, func()) {
	t.Helper()

	var advisoryLocker postgres.AdvisoryLocker = &postgres.NullAdvisoryLocker{}
	for _, flag := range flagsAndDeps {
		switch dep := flag.(type) {
		case postgres.AdvisoryLocker:
			advisoryLocker = dep
		}
	}
	s, err := strpkg.NewInsecureStore(c, advisoryLocker, gracefulpanic.NewSignal())
	if err != nil {
		require.NoError(t, err)
	}
	s.Config.SetDB(s.DB)
	return s, func() {
		cleanUpStore(t, s)
	}
}

// NewStore creates a new store
func NewStore(t *testing.T, flagsAndDeps ...interface{}) (*strpkg.Store, func()) {
	t.Helper()

	c := NewTestEVMConfig(t)
	store, storeCleanup := NewStoreWithConfig(t, c, flagsAndDeps...)
	return store, storeCleanup
}

func NewKeyStore(t testing.TB, db *gorm.DB) *keystore.Master {
	return keystore.New(db, utils.FastScryptParams)
}

func cleanUpStore(t testing.TB, store *strpkg.Store) {
	t.Helper()

	defer func() {
		if err := os.RemoveAll(store.Config.RootDir()); err != nil {
			logger.Warn("unable to clear test store:", err)
		}
	}()
	// Ignore sync errors for testing
	_ = logger.Sync()
	require.NoError(t, store.Close())
}

func ParseJSON(t testing.TB, body io.Reader) models.JSON {
	t.Helper()

	b, err := ioutil.ReadAll(body)
	require.NoError(t, err)
	return models.JSON{Result: gjson.ParseBytes(b)}
}

func ParseJSONAPIErrors(t testing.TB, body io.Reader) *models.JSONAPIErrors {
	t.Helper()

	b, err := ioutil.ReadAll(body)
	require.NoError(t, err)
	var respJSON models.JSONAPIErrors
	err = json.Unmarshal(b, &respJSON)
	require.NoError(t, err)
	return &respJSON
}

// MustReadFile loads a file but should never fail
func MustReadFile(t testing.TB, file string) []byte {
	t.Helper()

	content, err := ioutil.ReadFile(file)
	require.NoError(t, err)
	return content
}

type HTTPClientCleaner struct {
	HTTPClient cmd.HTTPClient
	t          testing.TB
}

func (r *HTTPClientCleaner) Get(path string, headers ...map[string]string) (*http.Response, func()) {
	resp, err := r.HTTPClient.Get(path, headers...)
	return bodyCleaner(r.t, resp, err)
}

func (r *HTTPClientCleaner) Post(path string, body io.Reader) (*http.Response, func()) {
	resp, err := r.HTTPClient.Post(path, body)
	return bodyCleaner(r.t, resp, err)
}

func (r *HTTPClientCleaner) Put(path string, body io.Reader) (*http.Response, func()) {
	resp, err := r.HTTPClient.Put(path, body)
	return bodyCleaner(r.t, resp, err)
}

func (r *HTTPClientCleaner) Patch(path string, body io.Reader, headers ...map[string]string) (*http.Response, func()) {
	resp, err := r.HTTPClient.Patch(path, body, headers...)
	return bodyCleaner(r.t, resp, err)
}

func (r *HTTPClientCleaner) Delete(path string) (*http.Response, func()) {
	resp, err := r.HTTPClient.Delete(path)
	return bodyCleaner(r.t, resp, err)
}

func bodyCleaner(t testing.TB, resp *http.Response, err error) (*http.Response, func()) {
	t.Helper()

	require.NoError(t, err)
	return resp, func() { require.NoError(t, resp.Body.Close()) }
}

// ParseResponseBody will parse the given response into a byte slice
func ParseResponseBody(t testing.TB, resp *http.Response) []byte {
	t.Helper()

	b, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	return b
}

// ParseJSONAPIResponse parses the response and returns the JSONAPI resource.
func ParseJSONAPIResponse(t testing.TB, resp *http.Response, resource interface{}) error {
	t.Helper()

	input := ParseResponseBody(t, resp)
	err := jsonapi.Unmarshal(input, resource)
	if err != nil {
		return fmt.Errorf("web: unable to unmarshal data, %+v", err)
	}

	return nil
}

// ParseJSONAPIResponseMeta parses the bytes of the root document and returns a
// map of *json.RawMessage's within the 'meta' key.
func ParseJSONAPIResponseMeta(input []byte) (map[string]*json.RawMessage, error) {
	var root map[string]*json.RawMessage
	err := json.Unmarshal(input, &root)
	if err != nil {
		return root, err
	}

	var meta map[string]*json.RawMessage
	err = json.Unmarshal(*root["meta"], &meta)
	return meta, err
}

// ParseJSONAPIResponseMetaCount parses the bytes of the root document and
// returns the value of the 'count' key from the 'meta' section.
func ParseJSONAPIResponseMetaCount(input []byte) (int, error) {
	meta, err := ParseJSONAPIResponseMeta(input)
	if err != nil {
		return -1, err
	}

	var metaCount int
	err = json.Unmarshal(*meta["count"], &metaCount)
	return metaCount, err
}

// ReadLogs returns the contents of the applications log file as a string
func ReadLogs(cfg config.GeneralConfig) (string, error) {
	logFile := fmt.Sprintf("%s/log.jsonl", cfg.RootDir())
	b, err := ioutil.ReadFile(logFile)
	return string(b), err
}

func CreateJobViaWeb(t testing.TB, app *TestApplication, request []byte) job.Job {
	t.Helper()

	client := app.NewHTTPClient()
	resp, cleanup := client.Post("/v2/jobs", bytes.NewBuffer(request))
	defer cleanup()
	AssertServerResponse(t, resp, http.StatusOK)

	var createdJob job.Job
	err := ParseJSONAPIResponse(t, resp, &createdJob)
	require.NoError(t, err)
	return createdJob
}

func CreateJobViaWeb2(t testing.TB, app *TestApplication, spec string) webpresenters.JobResource {
	t.Helper()

	client := app.NewHTTPClient()
	resp, cleanup := client.Post("/v2/jobs", bytes.NewBufferString(spec))
	defer cleanup()
	AssertServerResponse(t, resp, http.StatusOK)

	var jobResponse webpresenters.JobResource
	err := ParseJSONAPIResponse(t, resp, &jobResponse)
	require.NoError(t, err)
	return jobResponse
}

func DeleteJobViaWeb(t testing.TB, app *TestApplication, jobID int32) {
	t.Helper()

	client := app.NewHTTPClient()
	resp, cleanup := client.Delete(fmt.Sprintf("/v2/jobs/%v", jobID))
	defer cleanup()
	AssertServerResponse(t, resp, http.StatusNoContent)
}

func AwaitJobActive(t testing.TB, jobSpawner job.Spawner, jobID int32, waitFor time.Duration) {
	t.Helper()
	require.Eventually(t, func() bool {
		_, exists := jobSpawner.ActiveJobs()[jobID]
		return exists
	}, waitFor, 10*time.Millisecond)
}

func CreateJobRunViaExternalInitiatorV2(
	t testing.TB,
	app *TestApplication,
	jobID uuid.UUID,
	eia auth.Token,
	body string,
) webpresenters.PipelineRunResource {
	t.Helper()

	headers := make(map[string]string)
	headers[static.ExternalInitiatorAccessKeyHeader] = eia.AccessKey
	headers[static.ExternalInitiatorSecretHeader] = eia.Secret

	url := app.Config.ClientNodeURL() + "/v2/jobs/" + jobID.String() + "/runs"
	bodyBuf := bytes.NewBufferString(body)
	resp, cleanup := UnauthenticatedPost(t, url, bodyBuf, headers)
	defer cleanup()
	AssertServerResponse(t, resp, 200)
	var pr webpresenters.PipelineRunResource
	err := ParseJSONAPIResponse(t, resp, &pr)
	require.NoError(t, err)

	// assert.Equal(t, j.ID, pr.JobSpecID)
	return pr
}

// CreateBridgeTypeViaWeb creates a bridgetype via web using /v2/bridge_types
func CreateBridgeTypeViaWeb(
	t testing.TB,
	app *TestApplication,
	payload string,
) *webpresenters.BridgeResource {
	t.Helper()

	client := app.NewHTTPClient()
	resp, cleanup := client.Post(
		"/v2/bridge_types",
		bytes.NewBufferString(payload),
	)
	defer cleanup()
	AssertServerResponse(t, resp, http.StatusOK)
	bt := &webpresenters.BridgeResource{}
	err := ParseJSONAPIResponse(t, resp, bt)
	require.NoError(t, err)

	return bt
}

// CreateExternalInitiatorViaWeb creates a bridgetype via web using /v2/bridge_types
func CreateExternalInitiatorViaWeb(
	t testing.TB,
	app *TestApplication,
	payload string,
) *webpresenters.ExternalInitiatorAuthentication {
	t.Helper()

	client := app.NewHTTPClient()
	resp, cleanup := client.Post(
		"/v2/external_initiators",
		bytes.NewBufferString(payload),
	)
	defer cleanup()
	AssertServerResponse(t, resp, http.StatusCreated)
	ei := &webpresenters.ExternalInitiatorAuthentication{}
	err := ParseJSONAPIResponse(t, resp, ei)
	require.NoError(t, err)

	return ei
}

const (
	// DBWaitTimeout is how long we wait by default for something to appear in
	// the DB. It needs to be fairly long because integration
	// tests rely on it.
	DBWaitTimeout = 20 * time.Second
	// DBPollingInterval can't be too short to avoid DOSing the test database
	DBPollingInterval = 100 * time.Millisecond
	// AssertNoActionTimeout shouldn't be too long, or it will slow down tests
	AssertNoActionTimeout = 3 * time.Second
)

// WaitForSpecErrorV2 polls until the passed in jobID has count number
// of job spec errors.
func WaitForSpecErrorV2(t *testing.T, store *strpkg.Store, jobID int32, count int) []job.SpecError {
	t.Helper()

	g := gomega.NewGomegaWithT(t)
	var jse []job.SpecError
	g.Eventually(func() []job.SpecError {
		err := store.DB.
			Where("job_id = ?", jobID).
			Find(&jse).Error
		assert.NoError(t, err)
		return jse
	}, DBWaitTimeout, DBPollingInterval).Should(gomega.HaveLen(count))
	return jse
}

func WaitForPipelineRuns(t testing.TB, nodeID int, jobID int32, jo job.ORM, want int, timeout, poll time.Duration) []pipeline.Run {
	t.Helper()

	var err error
	prs := []pipeline.Run{}
	gomega.NewGomegaWithT(t).Eventually(func() []pipeline.Run {
		prs, _, err = jo.PipelineRunsByJobID(jobID, 0, 1000)
		assert.NoError(t, err)
		return prs
	}, timeout, poll).Should(gomega.HaveLen(want))

	return prs
}

func WaitForPipelineComplete(t testing.TB, nodeID int, jobID int32, expectedPipelineRuns int, expectedTaskRuns int, jo job.ORM, timeout, poll time.Duration) []pipeline.Run {
	t.Helper()

	var pr []pipeline.Run
	var numPipelineRuns int
	gomega.NewGomegaWithT(t).Eventually(func() bool {
		prs, _, err := jo.PipelineRunsByJobID(jobID, 0, 1000)
		assert.NoError(t, err)
		var completed []pipeline.Run

		for i := range prs {
			if prs[i].State == pipeline.RunStatusCompleted {
				if !prs[i].Outputs.Null {
					if !prs[i].Errors.HasError() {
						// txdb effectively ignores transactionality of queries, so we need to explicitly expect a number of task runs
						// (if the read occurrs mid-transaction and a job run in inserted but task runs not yet).
						if len(prs[i].PipelineTaskRuns) == expectedTaskRuns {
							completed = append(completed, prs[i])
						}
					}
				}
			}
		}
		numPipelineRuns = len(completed)
		if len(completed) >= expectedPipelineRuns {
			pr = completed
			return true
		}
		return false
	}, timeout, poll).Should(gomega.BeTrue(), fmt.Sprintf("job %d on node %d not complete with %d runs (found %v runs)", jobID, nodeID, expectedPipelineRuns, numPipelineRuns))
	return pr
}

// AssertPipelineRunsStays asserts that the number of pipeline runs for a particular job remains at the provided values
func AssertPipelineRunsStays(t testing.TB, pipelineSpecID int32, store *strpkg.Store, want int) []pipeline.Run {
	t.Helper()
	g := gomega.NewGomegaWithT(t)

	var prs []pipeline.Run
	g.Consistently(func() []pipeline.Run {
		err := store.DB.
			Where("pipeline_spec_id = ?", pipelineSpecID).
			Find(&prs).Error
		assert.NoError(t, err)
		return prs
	}, AssertNoActionTimeout, DBPollingInterval).Should(gomega.HaveLen(want))
	return prs
}

func WaitForEthTxAttemptsForEthTx(t testing.TB, store *strpkg.Store, ethTx bulletprooftxmanager.EthTx) []bulletprooftxmanager.EthTxAttempt {
	t.Helper()
	g := gomega.NewGomegaWithT(t)

	var attempts []bulletprooftxmanager.EthTxAttempt
	var err error
	g.Eventually(func() int {
		err = store.DB.Order("created_at desc").Where("eth_tx_id = ?", ethTx.ID).Find(&attempts).Error
		assert.NoError(t, err)
		return len(attempts)
	}, DBWaitTimeout, DBPollingInterval).Should(gomega.BeNumerically(">", 0))
	return attempts
}

func WaitForEthTxAttemptCount(t testing.TB, store *strpkg.Store, want int) []bulletprooftxmanager.EthTxAttempt {
	t.Helper()
	g := gomega.NewGomegaWithT(t)

	var txas []bulletprooftxmanager.EthTxAttempt
	var err error
	g.Eventually(func() []bulletprooftxmanager.EthTxAttempt {
		err = store.DB.Find(&txas).Error
		assert.NoError(t, err)
		return txas
	}, DBWaitTimeout, DBPollingInterval).Should(gomega.HaveLen(want))
	return txas
}

// AssertEthTxAttemptCountStays asserts that the number of tx attempts remains at the provided value
func AssertEthTxAttemptCountStays(t testing.TB, store *strpkg.Store, want int) []bulletprooftxmanager.EthTxAttempt {
	t.Helper()
	g := gomega.NewGomegaWithT(t)

	var txas []bulletprooftxmanager.EthTxAttempt
	var err error
	g.Consistently(func() []bulletprooftxmanager.EthTxAttempt {
		err = store.DB.Find(&txas).Error
		assert.NoError(t, err)
		return txas
	}, AssertNoActionTimeout, DBPollingInterval).Should(gomega.HaveLen(want))
	return txas
}

// ParseISO8601 given the time string it Must parse the time and return it
func ParseISO8601(t testing.TB, s string) time.Time {
	t.Helper()

	tm, err := time.Parse(time.RFC3339Nano, s)
	require.NoError(t, err)
	return tm
}

// NullableTime will return a valid nullable time given time.Time
func NullableTime(t time.Time) null.Time {
	return null.TimeFrom(t)
}

// ParseNullableTime given a time string parse it into a null.Time
func ParseNullableTime(t testing.TB, s string) null.Time {
	t.Helper()

	return NullableTime(ParseISO8601(t, s))
}

// Head given the value convert it into an Head
func Head(val interface{}) *models.Head {
	var h models.Head
	time := uint64(0)
	switch t := val.(type) {
	case int:
		h = models.NewHead(big.NewInt(int64(t)), utils.NewHash(), utils.NewHash(), time)
	case uint64:
		h = models.NewHead(big.NewInt(int64(t)), utils.NewHash(), utils.NewHash(), time)
	case int64:
		h = models.NewHead(big.NewInt(t), utils.NewHash(), utils.NewHash(), time)
	case *big.Int:
		h = models.NewHead(t, utils.NewHash(), utils.NewHash(), time)
	default:
		logger.Panicf("Could not convert %v of type %T to Head", val, val)
	}
	return &h
}

// TransactionsFromGasPrices returns transactions matching the given gas prices
func TransactionsFromGasPrices(gasPrices ...int64) []gas.Transaction {
	txs := make([]gas.Transaction, len(gasPrices))
	for i, gasPrice := range gasPrices {
		txs[i] = gas.Transaction{GasPrice: big.NewInt(gasPrice), GasLimit: 42}
	}
	return txs
}

// BlockWithTransactions returns a new ethereum block with transactions
// matching the given gas prices
func BlockWithTransactions(gasPrices ...int64) *types.Block {
	txs := make([]*types.Transaction, len(gasPrices))
	for i, gasPrice := range gasPrices {
		txs[i] = types.NewTransaction(0, common.Address{}, nil, 0, big.NewInt(gasPrice), nil)
	}
	return types.NewBlock(&types.Header{}, txs, nil, nil, new(trie.Trie))
}

type TransactionReceipter interface {
	TransactionReceipt(context.Context, common.Hash) (*types.Receipt, error)
}

func RequireTxSuccessful(t testing.TB, client TransactionReceipter, txHash common.Hash) *types.Receipt {
	t.Helper()
	r, err := client.TransactionReceipt(context.Background(), txHash)
	require.NoError(t, err)
	require.NotNil(t, r)
	require.Equal(t, uint64(1), r.Status)
	return r
}

func StringToHash(s string) common.Hash {
	return common.BytesToHash([]byte(s))
}

// AssertServerResponse is used to match against a client response, will print
// any errors returned if the request fails.
func AssertServerResponse(t testing.TB, resp *http.Response, expectedStatusCode int) {
	t.Helper()

	if resp.StatusCode == expectedStatusCode {
		return
	}

	t.Logf("expected status code %s got %s", http.StatusText(expectedStatusCode), http.StatusText(resp.StatusCode))

	if resp.StatusCode >= 300 && resp.StatusCode < 600 {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			assert.FailNowf(t, "Unable to read body", err.Error())
		}

		var result map[string][]string
		err = json.Unmarshal(b, &result)
		if err != nil {
			assert.FailNowf(t, fmt.Sprintf("Unable to unmarshal json from body '%s'", string(b)), err.Error())
		}

		assert.FailNowf(t, "Request failed", "Expected %d response, got %d with errors: %s", expectedStatusCode, resp.StatusCode, result["errors"])
	} else {
		assert.FailNowf(t, "Unexpected response", "Expected %d response, got %d", expectedStatusCode, resp.StatusCode)
	}
}

func DecodeSessionCookie(value string) (string, error) {
	var decrypted map[interface{}]interface{}
	codecs := securecookie.CodecsFromPairs([]byte(SessionSecret))
	err := securecookie.DecodeMulti(web.SessionName, value, &decrypted, codecs...)
	if err != nil {
		return "", err
	}
	value, ok := decrypted[web.SessionIDKey].(string)
	if !ok {
		return "", fmt.Errorf("decrypted[web.SessionIDKey] is not a string (%v)", value)
	}
	return value, nil
}

func MustGenerateSessionCookie(value string) *http.Cookie {
	decrypted := map[interface{}]interface{}{web.SessionIDKey: value}
	codecs := securecookie.CodecsFromPairs([]byte(SessionSecret))
	encoded, err := securecookie.EncodeMulti(web.SessionName, decrypted, codecs...)
	if err != nil {
		logger.Panic(err)
	}
	return sessions.NewCookie(web.SessionName, encoded, &sessions.Options{})
}

func NormalizedJSON(t testing.TB, input []byte) string {
	t.Helper()

	normalized, err := utils.NormalizedJSON(input)
	require.NoError(t, err)
	return normalized
}

func AssertError(t testing.TB, want bool, err error) {
	t.Helper()

	if want {
		assert.Error(t, err)
	} else {
		assert.NoError(t, err)
	}
}

func UnauthenticatedPost(t testing.TB, url string, body io.Reader, headers map[string]string) (*http.Response, func()) {
	t.Helper()

	client := http.Client{}
	request, err := http.NewRequest("POST", url, body)
	require.NoError(t, err)
	request.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		request.Header.Add(key, value)
	}
	resp, err := client.Do(request)
	require.NoError(t, err)
	return resp, func() { resp.Body.Close() }
}

func UnauthenticatedPatch(t testing.TB, url string, body io.Reader, headers map[string]string) (*http.Response, func()) {
	t.Helper()

	client := http.Client{}
	request, err := http.NewRequest("PATCH", url, body)
	require.NoError(t, err)
	request.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		request.Header.Add(key, value)
	}
	resp, err := client.Do(request)
	require.NoError(t, err)
	return resp, func() { resp.Body.Close() }
}

func MustParseDuration(t testing.TB, durationStr string) time.Duration {
	t.Helper()

	duration, err := time.ParseDuration(durationStr)
	require.NoError(t, err)
	return duration
}

func NewSession(optionalSessionID ...string) models.Session {
	session := models.NewSession()
	if len(optionalSessionID) > 0 {
		session.ID = optionalSessionID[0]
	}
	return session
}

func AllExternalInitiators(t testing.TB, store *strpkg.Store) []models.ExternalInitiator {
	t.Helper()

	var all []models.ExternalInitiator
	err := store.RawDBWithAdvisoryLock(func(db *gorm.DB) error {
		return db.Find(&all).Error
	})
	require.NoError(t, err)
	return all
}

func GetLastEthTxAttempt(t testing.TB, store *strpkg.Store) bulletprooftxmanager.EthTxAttempt {
	t.Helper()

	var txa bulletprooftxmanager.EthTxAttempt
	var count int64
	err := store.ORM.RawDBWithAdvisoryLock(func(db *gorm.DB) error {
		return db.Order("created_at desc").First(&txa).Count(&count).Error
	})
	require.NoError(t, err)
	require.NotEqual(t, 0, count)
	return txa
}

type Awaiter chan struct{}

func NewAwaiter() Awaiter { return make(Awaiter) }

func (a Awaiter) ItHappened() { close(a) }

func (a Awaiter) AssertHappened(t *testing.T, expected bool) {
	t.Helper()
	select {
	case <-a:
		if !expected {
			t.Fatal("It happened")
		}
	default:
		if expected {
			t.Fatal("It didn't happen")
		}
	}
}

func (a Awaiter) AwaitOrFail(t testing.TB, durationParams ...time.Duration) {
	t.Helper()

	duration := 10 * time.Second
	if len(durationParams) > 0 {
		duration = durationParams[0]
	}

	select {
	case <-a:
	case <-time.After(duration):
		t.Fatal("Timed out waiting for Awaiter to get ItHappened")
	}
}

func CallbackOrTimeout(t testing.TB, msg string, callback func(), durationParams ...time.Duration) {
	t.Helper()

	duration := 100 * time.Millisecond
	if len(durationParams) > 0 {
		duration = durationParams[0]
	}

	done := make(chan struct{})
	go func() {
		callback()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(duration):
		t.Fatal(fmt.Sprintf("CallbackOrTimeout: %s timed out", msg))
	}
}

func MustParseURL(input string) *url.URL {
	u, err := url.Parse(input)
	if err != nil {
		logger.Panic(err)
	}
	return u
}

// GenericEncode eth encodes values based on the provided types
func GenericEncode(types []string, values ...interface{}) ([]byte, error) {
	if len(values) != len(types) {
		return nil, errors.New("must include same number of values as types")
	}
	var args abi.Arguments
	for _, t := range types {
		ty, _ := abi.NewType(t, "", nil)
		args = append(args, abi.Argument{Type: ty})
	}
	out, err := args.PackValues(values)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func MustGenericEncode(types []string, values ...interface{}) []byte {
	if len(values) != len(types) {
		panic("must include same number of values as types")
	}
	var args abi.Arguments
	for _, t := range types {
		ty, _ := abi.NewType(t, "", nil)
		args = append(args, abi.Argument{Type: ty})
	}
	out, err := args.PackValues(values)
	if err != nil {
		panic(err)
	}
	return out
}

func MakeRoundStateReturnData(
	roundID uint64,
	eligible bool,
	answer, startAt, timeout, availableFunds, paymentAmount, oracleCount uint64,
) []byte {
	var data []byte
	if eligible {
		data = append(data, utils.EVMWordUint64(1)...)
	} else {
		data = append(data, utils.EVMWordUint64(0)...)
	}
	data = append(data, utils.EVMWordUint64(roundID)...)
	data = append(data, utils.EVMWordUint64(answer)...)
	data = append(data, utils.EVMWordUint64(startAt)...)
	data = append(data, utils.EVMWordUint64(timeout)...)
	data = append(data, utils.EVMWordUint64(availableFunds)...)
	data = append(data, utils.EVMWordUint64(oracleCount)...)
	data = append(data, utils.EVMWordUint64(paymentAmount)...)
	return data
}

var fluxAggregatorABI = eth.MustGetABI(flux_aggregator_wrapper.FluxAggregatorABI)

func MockFluxAggCall(client *mocks.Client, address common.Address, funcName string) *mock.Call {
	funcSig := hexutil.Encode(fluxAggregatorABI.Methods[funcName].ID)
	if len(funcSig) != 10 {
		panic(fmt.Sprintf("Unable to find FluxAgg function with name %s", funcName))
	}
	return client.On(
		"CallContract",
		mock.Anything,
		mock.MatchedBy(func(callArgs ethereum.CallMsg) bool {
			return *callArgs.To == address &&
				hexutil.Encode(callArgs.Data)[0:10] == funcSig
		}),
		mock.Anything)
}

// EthereumLogIterator is the interface provided by gethwrapper representations of EVM
// logs.
type EthereumLogIterator interface{ Next() bool }

// GetLogs drains logs of EVM log representations. Since those log
// representations don't fit into a type hierarchy, this API is a bit awkward.
// It returns the logs as a slice of blank interface{}s, and if rv is non-nil,
// it must be a pointer to a slice for elements of the same type as the logs,
// in which case GetLogs will append the logs to it.
func GetLogs(t *testing.T, rv interface{}, logs EthereumLogIterator) []interface{} {
	v := reflect.ValueOf(rv)
	require.True(t, rv == nil ||
		v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Slice,
		"must pass a slice to receive logs")
	var e reflect.Value
	if rv != nil {
		e = v.Elem()
	}
	var irv []interface{}
	for logs.Next() {
		log := reflect.Indirect(reflect.ValueOf(logs)).FieldByName("Event")
		if v.Kind() == reflect.Ptr {
			e.Set(reflect.Append(e, log))
		}
		irv = append(irv, log.Interface())
	}
	return irv
}

func MakeConfigDigest(t *testing.T) ocrtypes.ConfigDigest {
	t.Helper()
	b := make([]byte, 16)
	/* #nosec G404 */
	_, err := rand.Read(b)
	if err != nil {
		t.Fatal(err)
	}
	return MustBytesToConfigDigest(t, b)
}

func MustBytesToConfigDigest(t *testing.T, b []byte) ocrtypes.ConfigDigest {
	t.Helper()
	configDigest, err := ocrtypes.BytesToConfigDigest(b)
	if err != nil {
		t.Fatal(err)
	}
	return configDigest
}

// MockApplicationEthCalls mocks all calls made by the chainlink application as
// standard when starting and stopping
func MockApplicationEthCalls(t *testing.T, app *TestApplication, ethClient *mocks.Client) (verify func()) {
	t.Helper()

	// Start
	ethClient.On("Dial", mock.Anything).Return(nil)
	sub := new(mocks.Subscription)
	sub.On("Err").Return(nil)
	ethClient.On("SubscribeNewHead", mock.Anything, mock.Anything).Return(sub, nil)
	ethClient.On("ChainID", mock.Anything).Return(app.Store.Config.ChainID(), nil)
	ethClient.On("PendingNonceAt", mock.Anything, mock.Anything).Return(uint64(0), nil).Maybe()
	ethClient.On("HeadByNumber", mock.Anything, mock.Anything).Return(nil, nil).Maybe()

	// Stop
	sub.On("Unsubscribe").Return(nil)

	return func() {
		ethClient.AssertExpectations(t)
	}
}

func MockSubscribeToLogsCh(ethClient *mocks.Client, sub *mocks.Subscription) chan chan<- types.Log {
	logsCh := make(chan chan<- types.Log, 1)
	ethClient.On("SubscribeFilterLogs", mock.Anything, mock.Anything, mock.Anything).
		Return(sub, nil).
		Run(func(args mock.Arguments) { // context.Context, ethereum.FilterQuery, chan<- types.Log
			logsCh <- args.Get(2).(chan<- types.Log)
		})
	return logsCh
}

func MustNewJSONSerializable(t *testing.T, s string) pipeline.JSONSerializable {
	t.Helper()

	js := new(pipeline.JSONSerializable)
	err := js.UnmarshalJSON([]byte(s))
	require.NoError(t, err)
	return *js
}

func BatchElemMatchesHash(req rpc.BatchElem, hash common.Hash) bool {
	return req.Method == "eth_getTransactionReceipt" &&
		len(req.Args) == 1 && req.Args[0] == hash
}

func BatchElemMustMatchHash(t *testing.T, req rpc.BatchElem, hash common.Hash) {
	t.Helper()
	if !BatchElemMatchesHash(req, hash) {
		t.Fatalf("Batch hash %v does not match expected %v", req.Args[0], hash)
	}
}

type SimulateIncomingHeadsArgs struct {
	StartBlock, EndBlock int64
	BackfillDepth        int64
	Interval             time.Duration
	Timeout              time.Duration
	HeadTrackables       []httypes.HeadTrackable
	Blocks               *Blocks
}

func SimulateIncomingHeads(t *testing.T, args SimulateIncomingHeadsArgs) (func(), chan struct{}) {
	t.Helper()
	logger.Infof("Simulating incoming heads from %v to %v...", args.StartBlock, args.EndBlock)

	if args.BackfillDepth == 0 {
		t.Fatal("BackfillDepth must be > 0")
	}

	// Build the full chain of heads
	heads := args.Blocks.Heads
	if args.Timeout == 0 {
		args.Timeout = 60 * time.Second
	}
	if args.Interval == 0 {
		args.Interval = 250 * time.Millisecond
	}
	ctx, cancel := context.WithTimeout(context.Background(), args.Timeout)
	defer cancel()
	chTimeout := time.After(args.Timeout)

	chDone := make(chan struct{})
	go func() {
		current := args.StartBlock
		for {
			select {
			case <-chDone:
				return
			case <-chTimeout:
				return
			default:
				_, exists := heads[current]
				if !exists {
					logger.Fatalf("Head %v does not exist", current)
				}

				for _, ht := range args.HeadTrackables {
					ht.OnNewLongestChain(ctx, *heads[current])
				}
				if args.EndBlock >= 0 && current == args.EndBlock {
					chDone <- struct{}{}
					return
				}
				current++
				time.Sleep(args.Interval)
			}
		}
	}()
	var once sync.Once
	cleanup := func() {
		once.Do(func() {
			close(chDone)
			cancel()
		})
	}
	return cleanup, chDone
}

type Blocks struct {
	t       *testing.T
	Hashes  []common.Hash
	mHashes map[int64]common.Hash
	Heads   map[int64]*models.Head
}

func (b *Blocks) LogOnBlockNum(i uint64, addr common.Address) types.Log {
	return RawNewRoundLog(b.t, addr, b.Hashes[i], i, 0, false)
}

func (b *Blocks) LogOnBlockNumRemoved(i uint64, addr common.Address) types.Log {
	return RawNewRoundLog(b.t, addr, b.Hashes[i], i, 0, true)
}

func (b *Blocks) LogOnBlockNumWithIndex(i uint64, logIndex uint, addr common.Address) types.Log {
	return RawNewRoundLog(b.t, addr, b.Hashes[i], i, logIndex, false)
}

func (b *Blocks) LogOnBlockNumWithIndexRemoved(i uint64, logIndex uint, addr common.Address) types.Log {
	return RawNewRoundLog(b.t, addr, b.Hashes[i], i, logIndex, true)
}

func (b *Blocks) LogOnBlockNumWithTopics(i uint64, logIndex uint, addr common.Address, topics []common.Hash) types.Log {
	return RawNewRoundLogWithTopics(b.t, addr, b.Hashes[i], i, logIndex, false, topics)
}

func (b *Blocks) HashesMap() map[int64]common.Hash {
	return b.mHashes
}

func (b *Blocks) Head(number uint64) *models.Head {
	return b.Heads[int64(number)]
}

func (b *Blocks) ForkAt(t *testing.T, blockNum int64, numHashes int) *Blocks {
	blocks2 := NewBlocks(t, len(b.Heads)+numHashes)

	if _, exists := blocks2.Heads[blockNum]; !exists {
		logger.Fatalf("Not enough length for block num: %v", blockNum)
	}
	blocks2.Heads[blockNum].Parent = b.Heads[blockNum].Parent
	return blocks2
}

func NewBlocks(t *testing.T, numHashes int) *Blocks {
	hashes := make([]common.Hash, 0)
	heads := make(map[int64]*models.Head)
	for i := int64(0); i < int64(numHashes); i++ {
		hash := utils.NewHash()
		hashes = append(hashes, hash)

		heads[i] = &models.Head{Hash: hash, Number: i}
		if i > 0 {
			parent := heads[i-1]
			heads[i].Parent = parent
			heads[i].ParentHash = parent.Hash
		}
	}

	hashesMap := make(map[int64]common.Hash)
	for i := 0; i < len(hashes); i++ {
		hashesMap[int64(i)] = hashes[i]
	}

	return &Blocks{
		t:       t,
		Hashes:  hashes,
		mHashes: hashesMap,
		Heads:   heads,
	}
}

type HeadTrackableFunc func(context.Context, models.Head)

func (fn HeadTrackableFunc) OnNewLongestChain(ctx context.Context, head models.Head) {
	fn(ctx, head)
}

type testifyExpectationsAsserter interface {
	AssertExpectations(t mock.TestingT) bool
}

type fakeT struct{}

func (ft fakeT) Logf(format string, args ...interface{})   {}
func (ft fakeT) Errorf(format string, args ...interface{}) {}
func (ft fakeT) FailNow()                                  {}

func EventuallyExpectationsMet(t *testing.T, mock testifyExpectationsAsserter, timeout time.Duration, interval time.Duration) {
	t.Helper()

	chTimeout := time.After(timeout)
	for {
		var ft fakeT
		success := mock.AssertExpectations(ft)
		if success {
			return
		}
		select {
		case <-chTimeout:
			mock.AssertExpectations(t)
			t.FailNow()
		default:
			time.Sleep(interval)
		}
	}
}

func AssertCount(t *testing.T, db *gorm.DB, model interface{}, expected int64) {
	t.Helper()
	var count int64
	err := db.Unscoped().Model(model).Count(&count).Error
	require.NoError(t, err)
	require.Equal(t, expected, count)
}

func WaitForCount(t testing.TB, store *strpkg.Store, model interface{}, want int64) {
	t.Helper()
	g := gomega.NewGomegaWithT(t)
	var count int64
	var err error
	g.Eventually(func() int64 {
		err = store.DB.Model(model).Count(&count).Error
		assert.NoError(t, err)
		return count
	}, DBWaitTimeout, DBPollingInterval).Should(gomega.Equal(want))
}

func AssertCountStays(t testing.TB, store *strpkg.Store, model interface{}, want int64) {
	t.Helper()
	g := gomega.NewGomegaWithT(t)
	var count int64
	var err error
	g.Consistently(func() int64 {
		err = store.DB.Model(model).Count(&count).Error
		assert.NoError(t, err)
		return count
	}, AssertNoActionTimeout, DBPollingInterval).Should(gomega.Equal(want))
}

func AssertRecordEventually(t *testing.T, store *strpkg.Store, model interface{}, check func() bool) {
	t.Helper()
	g := gomega.NewGomegaWithT(t)
	g.Eventually(func() bool {
		err := store.DB.Find(model).Error
		require.NoError(t, err, "unable to find record in DB")
		return check()
	}, DBWaitTimeout, DBPollingInterval).Should(gomega.BeTrue())
}

func MustSendingKeys(t *testing.T, ethKeyStore *keystore.Eth) (keys []ethkey.Key) {
	var err error
	keys, err = ethKeyStore.SendingKeys()
	require.NoError(t, err)
	return keys
}

func MustRandomP2PPeerID(t *testing.T) p2ppeer.ID {
	reader := rand.New(source)
	p2pPrivkey, _, err := cryptop2p.GenerateEd25519Key(reader)
	require.NoError(t, err)
	id, err := p2ppeer.IDFromPrivateKey(p2pPrivkey)
	require.NoError(t, err)
	return id
}

func MustWebURL(t *testing.T, s string) *models.WebURL {
	uri, err := url.Parse(s)
	require.NoError(t, err)
	return (*models.WebURL)(uri)
}

func AssertPipelineTaskRunsSuccessful(t testing.TB, runs []pipeline.TaskRun) {
	t.Helper()
	for i, run := range runs {
		require.True(t, run.Error.IsZero(), fmt.Sprintf("pipeline.Task run failed (idx: %v, dotID: %v, error: '%v')", i, run.GetDotID(), run.Error.ValueOrZero()))
	}
}
