/**
 * Copyright 2019, Z Lab Corporation. All rights reserved.
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/golang/glog"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/spf13/pflag"

	"github.com/zlabjp/openstack-iid-dynamic-json/pkg/endpoint/iid"
	"github.com/zlabjp/openstack-iid-dynamic-json/pkg/endpoint/iidkey"
	"github.com/zlabjp/openstack-iid-dynamic-json/pkg/keymanager"
)

var (
	flags = pflag.NewFlagSet("", pflag.ExitOnError)

	port              = flags.Int("port", 8080, `Port to listen to`)
	dataDir           = flags.String("dataDir", ".", `Directory name for saving data`)
	keyFile           = flags.String("keyPath", "keys.json", `Name of a key-pair file`)
	keyRotationPeriod = flags.String("keyRotationPeriod", "720h", `Period for key-pair rotation`)
	interval          = flags.String("interval", "5s", "Check key expiration with given interval")
)

const (
	keyType = keymanager.EC_P384
	Alg     = jwa.ES384
)

func init() {
	flag.Set("logtostderr", "true")
}

func main() {
	flags.AddGoFlagSet(flag.CommandLine)
	if err := flags.Parse(os.Args); err != nil {
		glog.Exitf("Failed to parse flags: %v", err)
	}

	if err := os.MkdirAll(*dataDir, 0755); err != nil {
		glog.Exitf("Failed to create dataDir: %v", err)
	}

	rp, err := time.ParseDuration(*keyRotationPeriod)
	if err != nil {
		glog.Exitf("Invalid Format for 'keyRotationPeriod', use Go-Style time duration value (e.g., 24h): %v", err)
	}
	iv, err := time.ParseDuration(*interval)
	if err != nil {
		glog.Exitf("Invalid Format for 'interval', use Go-Style time duration value (e.g., 24h): %v", err)
	}

	dm := keymanager.NewDiskKeyManager(keyType, filepath.Join(*dataDir, *keyFile), rp, iv)
	if err := dm.Initialize(); err != nil {
		glog.Exitf("Failed to initialize KeyManager: %v", err)
	}

	iidOpts := &iid.HandlerConfig{
		SignAlg:    Alg,
		KeyManager: dm,
	}
	http.HandleFunc("/iid", iidOpts.IIDHandler())

	iidKeyOpts := &iidkey.HandlerConfig{
		Alg:       Alg,
		KeyManger: dm,
	}
	http.HandleFunc("/iid_keys", iidKeyOpts.IIDKeyHandler())

	glog.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", *port), nil))
}
