package web

import (
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/chainlink/core/services/chainlink"
	"github.com/smartcontractkit/chainlink/core/web/presenters"
)

// OCRKeysController manages OCR key bundles
type OCRKeysController struct {
	App chainlink.Application
}

// Index lists OCR key bundles
// Example:
// "GET <application>/keys/ocr"
func (ocrkc *OCRKeysController) Index(c *gin.Context) {
	ekbs, err := ocrkc.App.GetKeyStore().OCR().GetOCRKeys()
	if err != nil {
		jsonAPIError(c, http.StatusInternalServerError, err)
		return
	}
	jsonAPIResponse(c, presenters.NewOCRKeysBundleResources(ekbs), "offChainReportingKeyBundle")
}

// Create and return an OCR key bundle
// Example:
// "POST <application>/keys/ocr"
func (ocrkc *OCRKeysController) Create(c *gin.Context) {
	key, err := ocrkc.App.GetKeyStore().OCR().GenerateOCRKey()
	if err != nil {
		jsonAPIError(c, http.StatusInternalServerError, err)
		return
	}
	jsonAPIResponse(c, presenters.NewOCRKeysBundleResource(key), "offChainReportingKeyBundle")
}

// Delete an OCR key bundle
// Example:
// "DELETE <application>/keys/ocr/:keyID"
// "DELETE <application>/keys/ocr/:keyID?hard=true"
func (ocrkc *OCRKeysController) Delete(c *gin.Context) {
	var err error
	id := c.Param("keyID")
	if err != nil {
		jsonAPIError(c, http.StatusUnprocessableEntity, err)
		return
	}
	ekb, err := ocrkc.App.GetKeyStore().OCR().GetOCRKey(id)
	if err != nil {
		jsonAPIError(c, http.StatusNotFound, err)
		return
	}
	err = ocrkc.App.GetKeyStore().OCR().DeleteOCRKey(id)
	if err != nil {
		jsonAPIError(c, http.StatusInternalServerError, err)
		return
	}
	jsonAPIResponse(c, presenters.NewOCRKeysBundleResource(ekb), "offChainReportingKeyBundle")
}

// Import imports an OCR key bundle
// Example:
// "Post <application>/keys/ocr/import"
func (ocrkc *OCRKeysController) Import(c *gin.Context) {
	defer logger.ErrorIfCalling(c.Request.Body.Close)

	bytes, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		jsonAPIError(c, http.StatusBadRequest, err)
		return
	}
	oldPassword := c.Query("oldpassword")
	encryptedOCRKeyBundle, err := ocrkc.App.GetKeyStore().OCR().ImportOCRKey(bytes, oldPassword)
	if err != nil {
		jsonAPIError(c, http.StatusInternalServerError, err)
		return
	}

	jsonAPIResponse(c, encryptedOCRKeyBundle, "offChainReportingKeyBundle")
}

// Export exports an OCR key bundle
// Example:
// "Post <application>/keys/ocr/export"
func (ocrkc *OCRKeysController) Export(c *gin.Context) {
	defer logger.ErrorIfCalling(c.Request.Body.Close)

	stringID := c.Param("ID")
	newPassword := c.Query("newpassword")
	bytes, err := ocrkc.App.GetKeyStore().OCR().ExportOCRKey(stringID, newPassword)
	if err != nil {
		jsonAPIError(c, http.StatusInternalServerError, err)
		return
	}

	c.Data(http.StatusOK, MediaType, bytes)
}
