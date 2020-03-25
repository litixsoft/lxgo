package lxPdf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// IPdf, interface for pdf service
type IPdf interface {
	CreatePdf(template string, data map[string]interface{}, opts ...*options) ([]byte, error)
}

type pdf struct {
	serviceUrl  string
	serviceKey  string
	templateDir string
}

func NewPdfService(serviceUrl, serviceKey, templateDir string) IPdf {
	return &pdf{
		serviceUrl:  serviceUrl,
		serviceKey:  serviceKey,
		templateDir: templateDir,
	}
}

func (p *pdf) CreatePdf(template string, data map[string]interface{}, opts ...*options) ([]byte, error) {
	if template == "" || data == nil {
		return nil, fmt.Errorf("invalid params")
	}

	// parse incoming map[string]interface to []byte
	content, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("malformed data: %v", err)
	}

	// Request body
	var requestBody bytes.Buffer

	// Create multipart writer
	multipartWriter := multipart.NewWriter(&requestBody)

	// Add template file to form body
	if err := p.addFormFile("template", template, multipartWriter); err != nil {
		return nil, fmt.Errorf("add template: %v", err)
	}

	// Add content data to form body
	if err := p.addFormField("data", content, multipartWriter); err != nil {
		return nil, fmt.Errorf("add data: %v", err)
	}

	// Merge options
	pdfOptions := MergeOptions(opts...)

	// add footer file to form body
	if pdfOptions.footer != nil {
		if err := p.addFormFile("footer", *pdfOptions.footer, multipartWriter); err != nil {
			log.Printf("add footer: %v", err)
		}

		if pdfOptions.footerData != nil {
			// parse incoming map[string]interface to []byte
			if footerData, err := json.Marshal(*pdfOptions.footerData); err == nil {
				if err := p.addFormField("footerData", footerData, multipartWriter); err != nil {
					log.Printf("add footer data: %v", err)
				}
			}
		}
	}

	// add header file to form body
	if pdfOptions.header != nil {
		if err := p.addFormFile("header", *pdfOptions.header, multipartWriter); err != nil {
			log.Printf("add header: %v", err)
		}

		if pdfOptions.headerData != nil {
			// parse incoming map[string]interface to []byte
			if headerData, err := json.Marshal(*pdfOptions.headerData); err == nil {
				if err := p.addFormField("headerData", headerData, multipartWriter); err != nil {
					log.Printf("add header data: %v", err)
				}
			}
		}
	}

	// add css files to form body
	if pdfOptions.css != nil {
		for _, css := range *pdfOptions.css {
			if err := p.addFormFile("css", css, multipartWriter); err != nil {
				log.Printf("add css: %v\n", err)
				continue
			}
		}
	}

	// add images files to form body
	if pdfOptions.images != nil {
		for _, image := range *pdfOptions.images {
			if err := p.addFormFile("image", image, multipartWriter); err != nil {
				log.Printf("add image: %v\n", err)
				continue
			}
		}
	}

	// add additional field to form body
	if pdfOptions.opts != nil {
		options, err := json.Marshal(pdfOptions.opts)
		if err == nil {
			if err := p.addFormField("options", options, multipartWriter); err != nil {
				return nil, fmt.Errorf("add data: %v", err)
			}
		}
	}

	// Close multipart writer to write the ending boundary
	if err := multipartWriter.Close(); err != nil {
		return nil, fmt.Errorf("writing ending boundary: %v", err)
	}

	// Build pdf service endpoint url
	u, err := url.Parse(p.serviceUrl)
	if err != nil {
		return nil, fmt.Errorf("invalid serviceUrl")
	}
	u.Path = filepath.Join(u.Path, "create")

	// Create request
	req, err := http.NewRequest("POST", u.String(), &requestBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %v", err)
	}

	// Add headers to request
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.serviceKey))
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())

	// Do the request
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %s", strings.ReplaceAll(err.Error(), "\"", ""))
	}

	if response.StatusCode == http.StatusOK {
		// Copy response to []byte
		result, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("copy response: %v", err)
		}
		return result, nil
	} else {
		responseStr := "unknown error"
		buf := new(bytes.Buffer)

		if _, err := buf.ReadFrom(response.Body); err == nil {
			responseStr = buf.String()
		}

		return nil, fmt.Errorf("service return with code %d: %s", response.StatusCode, responseStr)
	}
}

// addFormFile, append a file to multipart writer
func (p *pdf) addFormFile(field, filePath string, multipartWriter *multipart.Writer) error {
	_, filename := filepath.Split(filePath)

	// Open file
	file, err := os.Open(filepath.Join(p.templateDir, filePath))
	if err != nil {
		return fmt.Errorf("open file: %v", err)
	}
	defer func() { _ = file.Close() }()

	// Initialize the file field
	fileWriter, err := multipartWriter.CreateFormFile(field, filename)
	if err != nil {
		return fmt.Errorf("create form file: %v", err)
	}

	// Copy file content to filed writer
	if _, err := io.Copy(fileWriter, file); err != nil {
		return fmt.Errorf("copy file content: %v", err)
	}

	return nil
}

// addFormField, append form field to multipart writer
func (p *pdf) addFormField(field string, content []byte, multipartWriter *multipart.Writer) error {
	// Initialize the form field
	fieldWriter, err := multipartWriter.CreateFormField(field)
	if err != nil {
		return fmt.Errorf("create form field: %v", err)
	}

	// Copy content to form field
	if _, err := fieldWriter.Write(content); err != nil {
		return fmt.Errorf("write field content: %v", err)
	}

	return nil
}
