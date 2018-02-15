package renter

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"skybin/core"
	"skybin/metaserver"

	"github.com/gorilla/mux"
)

func NewServer(renter *Renter, logger *log.Logger) http.Handler {

	router := mux.NewRouter()

	server := &renterServer{
		renter: renter,
		logger: logger,
		router: router,
	}

	router.HandleFunc("/info", server.getInfo).Methods("GET")
	router.HandleFunc("/create-storage-estimate", server.createStorageEstimate).Methods("POST")
	router.HandleFunc("/reserve-storage", server.reserveStorage).Methods("POST")
	router.HandleFunc("/files", server.getFiles).Methods("GET")
	router.HandleFunc("/files/shared", server.getFiles).Methods("GET")
	router.HandleFunc("/files/upload", server.uploadFile).Methods("POST")
	router.HandleFunc("/files/download", server.downloadFile).Methods("POST")
	router.HandleFunc("/files/create-folder", server.createFolder).Methods("POST")
	router.HandleFunc("/files/share", server.shareFile).Methods("POST")
	router.HandleFunc("/files/rename", server.renameFile).Methods("POST")
	router.HandleFunc("/files/copy", server.copyFile).Methods("POST")
	router.HandleFunc("/files/remove", server.removeFile).Methods("POST")

	return server
}

type renterServer struct {
	renter *Renter
	logger *log.Logger
	router *mux.Router
	client *metaserver.Client
}

func (server *renterServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	server.logger.Println(r.RemoteAddr, r.Method, r.URL)
	server.router.ServeHTTP(w, r)
}

type errorResp struct {
	Error string `json:"error,omitempty"`
}

func (server *renterServer) getInfo(w http.ResponseWriter, r *http.Request) {
	info, err := server.renter.Info()
	if err != nil {
		server.writeResp(w, http.StatusInternalServerError,
			&errorResp{Error: err.Error()})
		return

	}
	server.writeResp(w, http.StatusOK, info)
}

func (server *renterServer) createStorageEstimate(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

type reserveStorageReq struct {
	Amount int64 `json:"amount"`
}

type reserveStorageResp struct {
	Contracts []*core.Contract `json:"contracts"`
}

func (server *renterServer) reserveStorage(w http.ResponseWriter, r *http.Request) {
	var req reserveStorageReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		server.logger.Println(err)
		server.writeResp(w, http.StatusBadRequest,
			&errorResp{Error: fmt.Sprintf("Unable to decode JSON. Error: %v", err)})
		return
	}

	contracts, err := server.renter.ReserveStorage(req.Amount)
	if err != nil {
		server.logger.Println(err)
		server.writeResp(w, http.StatusInternalServerError,
			&errorResp{Error: fmt.Sprintf("Unable to reserve storage. Error: %v", err)})
		return
	}

	resp := reserveStorageResp{
		Contracts: contracts,
	}
	server.writeResp(w, http.StatusCreated, &resp)
}

type getFilesResp struct {
	Files []*core.File `json:"files"`
}

func (server *renterServer) getFiles(w http.ResponseWriter, r *http.Request) {
	files, err := server.renter.ListFiles()
	if err != nil {
		server.writeResp(w, http.StatusInternalServerError,
			&errorResp{Error: err.Error()})
		return
	}
	server.writeResp(w, http.StatusOK, &getFilesResp{Files: files})
}

func (server *renterServer) getSharedFiles(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

type uploadFileReq struct {
	SourcePath      string `json:"sourcePath"`
	DestPath        string `json:"destPath"`
	ShouldOverwrite bool   `json:"shouldOverwrite"`
}

func (server *renterServer) uploadFile(w http.ResponseWriter, r *http.Request) {
	var req uploadFileReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		server.logger.Println(err)
		server.writeResp(w, http.StatusBadRequest,
			&errorResp{Error: fmt.Sprintf("Unable to decode JSON. Error: %v", err)})
		return
	}

	f, err := server.renter.Upload(req.SourcePath, req.DestPath, req.ShouldOverwrite)
	if err != nil {
		server.logger.Println(err)
		server.writeResp(w, http.StatusInternalServerError,
			&errorResp{Error: err.Error()})
		return
	}

	server.writeResp(w, http.StatusCreated, f)
}

type downloadFileReq struct {
	FileId   string `json:"fileId"`
	DestPath string `json:"destPath"`
}

func (server *renterServer) downloadFile(w http.ResponseWriter, r *http.Request) {
	var req downloadFileReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		server.logger.Println(err)
		server.writeResp(w, http.StatusBadRequest,
			&errorResp{Error: fmt.Sprintf("Unable to decode JSON. Error: %v", err)})
		return
	}

	err = server.renter.Download(req.FileId, req.DestPath)
	if err != nil {
		server.logger.Println(err)
		server.writeResp(w, http.StatusInternalServerError,
			&errorResp{Error: err.Error()})
		return
	}

	server.writeResp(w, http.StatusCreated, &errorResp{})
}

type createFolderReq struct {
	Name string `json:"name"`
}

func (server *renterServer) createFolder(w http.ResponseWriter, r *http.Request) {
	var req createFolderReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		server.logger.Println(err)
		server.writeResp(w, http.StatusBadRequest,
			&errorResp{Error: fmt.Sprintf("Unable to decode JSON. Error: %v", err)})
		return
	}

	f, err := server.renter.CreateFolder(req.Name)
	if err != nil {
		server.logger.Println(err)
		server.writeResp(w, http.StatusInternalServerError,
			&errorResp{Error: err.Error()})
		return
	}

	server.writeResp(w, http.StatusCreated, f)
}

type shareFileReq struct {
	FileId   string `json:"fileId"`
	RenterId string `json:"renterId"`
}

type shareFileResp struct {
	Message string `json:"message"`
}

func (server *renterServer) shareFile(w http.ResponseWriter, r *http.Request) {
	var req shareFileReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		server.logger.Println(err)
		server.writeResp(w, http.StatusBadRequest,
			&errorResp{Error: fmt.Sprintf("Unable to decode JSON. Error: %v", err)})
		return

	}

	if req.FileId == "" || req.RenterId == "" {
		server.writeResp(w, http.StatusBadRequest, &errorResp{Error: "must supply file ID and renter ID"})
		return
	}

	err = server.renter.ShareFile(req.FileId, req.RenterId)
	if err != nil {
		server.writeResp(w, http.StatusBadRequest, &errorResp{Error: err.Error()})
		return
	}

	server.writeResp(w, http.StatusOK, &shareFileResp{Message: "file shared"})
}

type renameFileReq struct {
	FileId string `json:"fileId"`
	Name   string `json:"name"`
}

func (server *renterServer) renameFile(w http.ResponseWriter, r *http.Request) {
	var req renameFileReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		server.logger.Println(err)
		server.writeResp(w, http.StatusBadRequest,
			&errorResp{Error: fmt.Sprintf("Unable to decode JSON. Error: %v", err)})
		return

	}

	f, err := server.renter.RenameFile(req.FileId, req.Name)
	if err != nil {
		server.logger.Println(err)
		server.writeResp(w, http.StatusBadRequest,
			&errorResp{Error: err.Error()})
		return
	}

	server.writeResp(w, http.StatusOK, f)
}

func (server *renterServer) copyFile(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
	w.WriteHeader(http.StatusNotImplemented)
}

type removeFileReq struct {
	FileID string `json:"fileID"`
}

func (server *renterServer) removeFile(w http.ResponseWriter, r *http.Request) {
	var req downloadFileReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		server.logger.Println(err)
		server.writeResp(w, http.StatusBadRequest,
			&errorResp{Error: fmt.Sprintf("Unable to decode JSON. Error: %v", err)})
		return
	}

	err = server.renter.Remove(req.FileId)
	if err != nil {
		server.logger.Println(err)
		server.writeResp(w, http.StatusInternalServerError,
			&errorResp{Error: err.Error()})
		return
	}

	server.writeResp(w, http.StatusOK, &errorResp{})
}

func (server *renterServer) writeResp(w http.ResponseWriter, status int, body interface{}) {
	w.WriteHeader(status)
	data, err := json.MarshalIndent(body, "", "    ")
	if err != nil {
		server.logger.Fatalf("error: cannot to encode response. error: %s", err)
	}
	_, err = w.Write(data)
	if err != nil {
		server.logger.Fatalf("error: cannot write response body. error: %s", err)
	}

	if r, ok := body.(*errorResp); ok && len(r.Error) > 0 {
		server.logger.Print(status, r)
	} else {
		server.logger.Println(status)
	}
}
