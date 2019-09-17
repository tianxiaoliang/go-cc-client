/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package servicecomb

import (
	"context"
	"errors"

	"github.com/apache/servicecomb-kie/client"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/go-chassis/go-chassis-config"
	"github.com/go-mesh/openlogging"
)

// Client contains the implementation of Client
type Client struct {
	KieClient     *client.Client
	DefaultLabels map[string]string
	opts          config.Options
}

const (
	//Name of the Plugin
	Name             = "servicecomb-kie"
	LabelService     = "serviceName"
	LabelVersion     = "version"
	LabelEnvironment = "environment"
	LabelApp         = "app"
)

// NewClient init the necessary objects needed for seamless communication to Kie Server
func NewClient(options config.Options) (config.Client, error) {
	kieClient := &Client{
		opts: options,
	}
	DefaultLabels := map[string]string{
		LabelApp:         options.App,
		LabelEnvironment: options.Env,
		LabelService:     options.ServiceName,
		LabelVersion:     options.Version,
	}
	configInfo := client.Config{Endpoint: kieClient.opts.ServerURI, DefaultLabels: DefaultLabels, VerifyPeer: kieClient.opts.EnableSSL}
	var err error
	kieClient.KieClient, err = client.New(configInfo)
	if err != nil {
		openlogging.Error("KieClient Initialization Failed: " + err.Error())
	}
	openlogging.Debug("KieClient Initialized successfully")
	return kieClient, err
}

// PullConfigs is used for pull config from servicecomb-kie
func (c *Client) PullConfigs(serviceName, version, app, env string) (map[string]interface{}, error) {
	openlogging.Debug("KieClient begin PullConfigs")
	labels := map[string]string{LabelService: serviceName, LabelVersion: version, LabelApp: app, LabelEnvironment: env}
	labelsAppLevel := map[string]string{LabelApp: app, LabelEnvironment: env}
	configsInfo := make(map[string]interface{})
	configurationsValue, err := c.KieClient.SearchByLabels(context.TODO(), client.WithGetProject(serviceName), client.WithLabels(labels, labelsAppLevel))
	if err != nil {
		openlogging.GetLogger().Errorf("Error in Querying the Response from Kie %s %#v", err.Error(), labels)
		return nil, err
	}
	openlogging.GetLogger().Debugf("KieClient SearchByLabels. %#v", labels)
	//Parse config result.
	for _, docRes := range configurationsValue {
		for _, docInfo := range docRes.Data {
			configsInfo[docInfo.Key] = docInfo.Value
		}
	}
	return configsInfo, nil
}

// PullConfig get config by key and labels.
func (c *Client) PullConfig(serviceName, version, app, env, key, contentType string) (interface{}, error) {
	labels := map[string]string{LabelService: serviceName, LabelVersion: version, LabelApp: app, LabelEnvironment: env}
	configurationsValue, err := c.KieClient.Get(context.TODO(), key, client.WithGetProject(serviceName), client.WithLabels(labels))
	if err != nil {
		openlogging.GetLogger().Error("Error in Querying the Response from Kie: " + err.Error())
		return nil, err
	}
	for _, doc := range configurationsValue {
		for _, kvDoc := range doc.Data {
			if key == kvDoc.Key {
				openlogging.GetLogger().Debugf("The Key Value of : ", kvDoc.Value)
				return doc, nil
			}
		}
	}
	return nil, errors.New("can not find value")
}

//PullConfigsByDI not implemented
func (c *Client) PullConfigsByDI(dimensionInfo string) (map[string]map[string]interface{}, error) {
	// TODO Return the configurations for customized Projects in Kie Configs
	return nil, errors.New("not implemented")
}

//PushConfigs put config in kie by key and labels.
func (c *Client) PushConfigs(data map[string]interface{}, serviceName, version, app, env string) (map[string]interface{}, error) {
	var configReq model.KVDoc
	labels := map[string]string{LabelService: serviceName, LabelVersion: version, LabelApp: app, LabelEnvironment: env}
	configResult := make(map[string]interface{})
	for key, configValue := range data {
		configReq.Key = key
		configReq.Value = configValue.(string)
		configReq.Labels = labels
		configurationsValue, err := c.KieClient.Put(context.TODO(), configReq, client.WithProject(serviceName))
		if err != nil {
			openlogging.Error("Error in PushConfigs to Kie: " + err.Error())
			return nil, err
		}
		openlogging.Debug("The Key Value of : " + configurationsValue.Value)
		configResult[configurationsValue.Key] = configurationsValue.Value
	}
	return configResult, nil
}

//DeleteConfigsByKeys use keyId for delete
func (c *Client) DeleteConfigsByKeys(keys []string, serviceName, version, app, env string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for _, keyId := range keys {
		err := c.KieClient.Delete(context.TODO(), keyId, "", client.WithProject(serviceName))
		if err != nil {
			openlogging.Error("Error in Delete from Kie. " + err.Error())
			return nil, err
		}
		openlogging.GetLogger().Debugf("Delete The KeyId:%s", keyId)
	}
	return result, nil
}

//Watch not implemented because kie not support.
func (c *Client) Watch(f func(map[string]interface{}), errHandler func(err error)) error {
	// TODO watch change events
	return errors.New("not implemented")
}

//Options.
func (c *Client) Options() config.Options {
	return c.opts
}

func init() {
	config.InstallConfigClientPlugin(Name, NewClient)
}
