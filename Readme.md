# Go-idmatch
Go-idmatch is a golang implementation of [idmatch](github.com/maddevsio/idmatch) project - opensource text recognition service. Currently, go-idmatch supports recognition of Kyrgyz Republic ID cards.

Demo website: [idmatch.co](https://idmatch.co/) 

### Prerequisites      
 Go-idmatch requires Gocv to be [installed](https://github.com/hybridgroup/gocv#how-to-install) in the system along with Tesseract OCR [language files](https://github.com/tesseract-ocr/tesseract/wiki/Data-Files) for English, Kyrgyz and Russian.
 
**Installation**
```
go get -u -d github.com/maddevsio/go-idmatch
cd $GOPATH/src/github.com/maddevsio/go-idmatch
go run main.go service
```
**Docker**
```
docker pull maddevsio/go-idmatch
docker run -p 8080:8080 -it --rm --name go-idmatch-service go-idmatch
```
**CLI**
```
./go-idmatch service				- starts web service on :8080
./go-idmatch ocr image <path>		- pass image to OCR engine	
```

