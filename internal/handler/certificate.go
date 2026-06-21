package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/andruho/courses/internal/domain"
	"github.com/andruho/courses/internal/service"
)

type CertificateHandler struct {
	certificates service.CertificateService
}

func NewCertificateHandler(certificates service.CertificateService) *CertificateHandler {
	return &CertificateHandler{certificates: certificates}
}

func (h *CertificateHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())

	certs, err := h.certificates.ListByUser(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"data": certs})
}

func (h *CertificateHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	certID := chi.URLParam(r, "id")

	cert, err := h.certificates.GetByID(r.Context(), userID, certID)
	if err != nil {
		writeError(w, err)
		return
	}
	if cert == nil {
		writeError(w, domain.ErrCertificateNotFound)
		return
	}
	writeOK(w, cert)
}
