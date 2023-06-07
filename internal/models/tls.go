package models

type TLS struct {
	CertFile   *string
	CertConfig *CertConfig
	ServerName *string
}

type CertConfig struct {
	Cert string
	Key  string
}
