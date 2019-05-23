/**
 * Copyright 2019, Z Lab Corporation. All rights reserved.
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package iidkey

import (
	"encoding/json"
	"net/http"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"

	"github.com/zlabjp/openstack-iid-dynamic-json/pkg/keymanager"
)

type HandlerConfig struct {
	Alg       jwa.SignatureAlgorithm
	KeyManger keymanager.KeyManager
}

func (c *HandlerConfig) IIDKeyHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		pubKeys, err := c.KeyManger.GetPublicKeys()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var keys []jwk.Key

		for i, p := range pubKeys {
			key, err := jwk.New(p)
			if err := key.Set(jwk.KeyIDKey, i); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			if err := key.Set(jwk.KeyUsageKey, string(jwk.ForSignature)); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			if err := key.Set(jwk.AlgorithmKey, c.Alg.String()); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			keys = append(keys, key)
		}

		jwkSet := jwk.Set{
			Keys: keys,
		}

		jsonbuf, err := json.MarshalIndent(jwkSet, "", "  ")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Write(jsonbuf)
	}
}
