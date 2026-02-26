package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strconv"
	"time"
)

type Auth struct {
	ak      string
	sk      string
	agentID string
}

func New(ak, sk, agentID string) *Auth {
	return &Auth{
		ak:      ak,
		sk:      sk,
		agentID: agentID,
	}
}

type Credentials struct {
	AK        string
	Timestamp int64
	Signature string
}

func (a *Auth) GenerateCredentials() *Credentials {
	ts := time.Now().UnixMilli()
	return &Credentials{
		AK:        a.ak,
		Timestamp: ts,
		Signature: a.generateSignature(ts),
	}
}

func (a *Auth) generateSignature(timestamp int64) string {
	h := hmac.New(sha256.New, []byte(a.sk))
	h.Write([]byte(strconv.FormatInt(timestamp, 10)))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func (a *Auth) Verify(creds *Credentials) bool {
	expected := a.generateSignature(creds.Timestamp)
	return creds.Signature == expected
}

func (a *Auth) Headers() map[string]string {
	ts := time.Now().UnixMilli()
	return map[string]string{
		"x-access-key": a.ak,
		"x-sign":       a.generateSignature(ts),
		"x-ts":         strconv.FormatInt(ts, 10),
		"x-agent-id":   a.agentID,
	}
}

func (a *Auth) AgentID() string {
	return a.agentID
}
