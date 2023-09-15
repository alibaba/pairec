package main

import "C"
import (
	"bytes"
	"fmt"
	"github.com/alibaba/pairec"
	"io/ioutil"
	"net/http/httptest"
	"os"
)

//export CommandRun
func CommandRun(configFile *C.char) bool {
	file := C.GoString(configFile)
	os.Setenv("CONFIG_PATH", file)
	os.Setenv("RUN_MODE", "COMMAND")
	pairec.Run()
	return true
}

//export Recommend
func Recommend(requestBody *C.char) *C.char {
	body := C.GoString(requestBody)
	readBuf := bytes.NewBufferString(body)
	req := httptest.NewRequest("POST", "/api/recall", readBuf)
	w := httptest.NewRecorder()
	pairec.PairecApp.Handlers.ServeHTTP(w, req)
	resp := w.Result()
	defer resp.Body.Close()
	responseBody, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(resp.StatusCode)
	fmt.Println(string(responseBody))

	return C.CString(string(responseBody))
}
func main() {
}
