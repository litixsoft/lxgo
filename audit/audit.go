package lxAudit

const (
	Insert = "insert"
	Update = "update"
	Delete = "delete"
)

// IAudit interface for lxaudit logger
type IAudit interface {
	Log(action string, user, data interface{}) error
}

// audit config struct
//type auditConfig struct {
//	client       *http.Client
//	clientHost   string
//	auditHost    string
//	auditAuthKey string
//}
//
//// audit struct
//type audit struct {
//	dbHost         string
//	dbName         string
//	collectionName string
//}
//
//var (
//	// mux for instance lock
//	auditMux = new(sync.Mutex)
//
//	// audit instance
//	auditConfigInstance *auditConfig
//)

// InitAuditInstance, set instance for auditConfig
//func InitAuditConfigInstance(clientHost, auditHost, auditAuthKey string) {
//	auditMux.Lock()
//	auditConfigInstance = &auditConfig{
//		client:       &http.Client{Timeout: time.Duration(10 * time.Second)},
//		clientHost:   clientHost,
//		auditHost:    auditHost,
//		auditAuthKey: auditAuthKey,
//	}
//	auditMux.Unlock()
//}

// GetAuditInstance, create new instance of audit with singleton auditConfig
//func GetAuditInstance(dbHost, dbName, collectionName string) IAudit {
//	// check config instance
//	if auditConfigInstance == nil {
//		panic(errors.New("auditConfigInstance was not initialized"))
//	}
//
//	// create audit instance
//	return &audit{
//		dbHost:         dbHost,
//		dbName:         dbName,
//		collectionName: collectionName,
//	}
//}

type audit struct {
	clientHost     string
	collectionName string
	auditHost      string
	auditAuthKey   string
}

func NewAudit(clientHost, collectionName, auditHost, auditAuthKey string) IAudit {
	return &audit{
		clientHost:     clientHost,
		collectionName: collectionName,
		auditHost:      auditHost,
		auditAuthKey:   auditAuthKey,
	}
}

// Log, send post request to audit service
func (a *audit) Log(action string, user, data interface{}) error {
	//	// Set entry for request
	//	entry := lxHelper.M{
	//		"host":       auditConfigInstance.clientHost,
	//		"db_host":    a.dbHost,
	//		"db":         a.dbName,
	//		"collection": a.collectionName,
	//		"action":     action,
	//		"user":       user,
	//		"data":       data,
	//	}
	//
	//	// Convert entry to json
	//	jsonData, err := json.Marshal(entry)
	//	if err != nil {
	//		return err
	//	}
	//
	//	// Post to url
	//	req, err := http.NewRequest("POST", auditConfigInstance.auditHost+"/log", bytes.NewBuffer(jsonData))
	//	if err != nil {
	//		return err
	//	}
	//
	//	// set header
	//	req.Header.Add("Authorization", "Bearer "+auditConfigInstance.auditAuthKey)
	//	req.Header.Add("Content-Type", "application/json")
	//
	//	// send request
	//	resp, err := auditConfigInstance.client.Do(req)
	//	if err != nil {
	//		return err
	//	}
	//
	//	// check validation or internal error
	//	if resp.StatusCode == http.StatusUnprocessableEntity || resp.StatusCode == http.StatusInternalServerError {
	//		body, err := ioutil.ReadAll(resp.Body)
	//		if err != nil {
	//			return err
	//		}
	//		if err := resp.Body.Close(); err != nil {
	//			return err
	//		}
	//
	//		// response error
	//		var respErr error
	//
	//		// check response error
	//		switch resp.StatusCode {
	//		case http.StatusInternalServerError:
	//			respErr = errors.New(fmt.Sprintf("Internal-Server-Error: %v", string(body)))
	//		case http.StatusUnprocessableEntity:
	//			respErr = errors.New(fmt.Sprintf("Validation-Error: %v", string(body)))
	//		}
	//
	//		return respErr
	//	}
	//
	//	return nil
	return nil
}
