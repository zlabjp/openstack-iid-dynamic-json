/**
 * Copyright 2019, Z Lab Corporation. All rights reserved.
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package iid

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jws"

	"github.com/zlabjp/openstack-iid-dynamic-json/pkg/keymanager"
	"github.com/zlabjp/openstack-iid-dynamic-json/pkg/vendordata"
)

type HandlerConfig struct {
	SignAlg    jwa.SignatureAlgorithm
	KeyManager keymanager.KeyManager
}

type Response struct {
	Data string `json:"data"`
}

type Payload struct {
	ID         string            `json:"jti"`
	ProjectID  string            `json:"projectID"`
	InstanceID string            `json:"instanceID"`
	ImageID    string            `json:"imageID"`
	Hostname   string            `json:"hostname"`
	Metadata   map[string]string `json:"metadata"`
	IssuedAt   int64             `json:"iat"`
	ExpiresAt  int64             `json:"exp"`
}

func (c *HandlerConfig) IIDHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		b, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		data := &vendordata.RequestData{}

		if err := json.Unmarshal(b, data); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		id, privKey, err := c.KeyManager.GetPrivateKey()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		now := time.Now()
		uuid := uuid.New().String()

		payload := &Payload{
			ID:         uuid,
			ProjectID:  data.ProjectID,
			InstanceID: data.InstanceID,
			ImageID:    data.ImageID,
			Hostname:   data.Hostname,
			Metadata:   data.Metadata,
			IssuedAt:   now.Unix(),
			ExpiresAt:  now.Add(10 * time.Minute).Unix(),
		}
		buf, err := json.Marshal(payload)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		h := &jws.StandardHeaders{
			JWStyp:   "JOSE+JSON",
			JWSkeyID: id,
		}
		jwsBuf, err := jws.Sign(buf, c.SignAlg, privKey, jws.WithHeaders(h))
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		resp := &Response{
			Data: string(jwsBuf),
		}

		result, err := json.MarshalIndent(resp, "", "    ")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Write(result)
	}
}
