/* Copyright © 2019 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func dataSourceNsxtPolicyGatewayQosProfile() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtPolicyGatewayQosProfileRead,

		Schema: map[string]*schema.Schema{
			"id":           getDataSourceIDSchema(),
			"display_name": getDataSourceDisplayNameSchema(),
			"description":  getDataSourceDescriptionSchema(),
			"path":         getPathSchema(),
		},
	}
}

func dataSourceNsxtPolicyGatewayQosProfileRead(d *schema.ResourceData, m interface{}) error {
	if isPolicyGlobalManager(m) {
		_, err := policyDataSourceResourceRead(d, getPolicyConnector(m), true, "GatewayQosProfile", nil)
		if err != nil {
			return err
		}
		return nil
	}
	connector := getPolicyConnector(m)
	client := infra.NewGatewayQosProfilesClient(connector)

	objID := d.Get("id").(string)
	objName := d.Get("display_name").(string)
	var obj model.GatewayQosProfile
	if objID != "" {
		// Get by id
		objGet, err := client.Get(objID)
		if err != nil {
			return handleDataSourceReadError(d, "GatewayQosProfile", objID, err)
		}
		obj = objGet
	} else if objName == "" {
		return fmt.Errorf("Error obtaining GatewayQosProfile ID or name during read")
	} else {
		// Get by full name/prefix
		includeMarkForDeleteObjectsParam := false
		objList, err := client.List(nil, &includeMarkForDeleteObjectsParam, nil, nil, nil, nil)
		if err != nil {
			return handleListError("GatewayQosProfile", err)
		}
		// go over the list to find the correct one (prefer a perfect match. If not - prefix match)
		var perfectMatch []model.GatewayQosProfile
		var prefixMatch []model.GatewayQosProfile
		for _, objInList := range objList.Results {
			if strings.HasPrefix(*objInList.DisplayName, objName) {
				prefixMatch = append(prefixMatch, objInList)
			}
			if *objInList.DisplayName == objName {
				perfectMatch = append(perfectMatch, objInList)
			}
		}
		if len(perfectMatch) > 0 {
			if len(perfectMatch) > 1 {
				return fmt.Errorf("Found multiple GatewayQosProfiles with name '%s'", objName)
			}
			obj = perfectMatch[0]
		} else if len(prefixMatch) > 0 {
			if len(prefixMatch) > 1 {
				return fmt.Errorf("Found multiple GatewayQosProfiles with name starting with '%s'", objName)
			}
			obj = prefixMatch[0]
		} else {
			return fmt.Errorf("GatewayQosProfile with name '%s' was not found", objName)
		}
	}

	d.SetId(*obj.Id)
	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	d.Set("path", obj.Path)
	return nil
}
