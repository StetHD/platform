// Copyright (c) 2017 Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package api4

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/mattermost/platform/model"
	"github.com/mattermost/platform/utils"
)

func TestCreateOAuthApp(t *testing.T) {
	th := Setup().InitBasic().InitSystemAdmin()
	defer TearDown()
	Client := th.Client
	AdminClient := th.SystemAdminClient

	enableOAuth := utils.Cfg.ServiceSettings.EnableOAuthServiceProvider
	adminOnly := *utils.Cfg.ServiceSettings.EnableOnlyAdminIntegrations
	defer func() {
		utils.Cfg.ServiceSettings.EnableOAuthServiceProvider = enableOAuth
		*utils.Cfg.ServiceSettings.EnableOnlyAdminIntegrations = adminOnly
	}()
	utils.Cfg.ServiceSettings.EnableOAuthServiceProvider = true
	utils.SetDefaultRolesBasedOnConfig()

	oapp := &model.OAuthApp{Name: GenerateTestAppName(), Homepage: "https://nowhere.com", Description: "test", CallbackUrls: []string{"https://nowhere.com"}}

	rapp, resp := AdminClient.CreateOAuthApp(oapp)
	CheckNoError(t, resp)

	if rapp.Name != oapp.Name {
		t.Fatal("names did not match")
	}

	*utils.Cfg.ServiceSettings.EnableOnlyAdminIntegrations = true
	utils.SetDefaultRolesBasedOnConfig()
	_, resp = Client.CreateOAuthApp(oapp)
	CheckForbiddenStatus(t, resp)

	*utils.Cfg.ServiceSettings.EnableOnlyAdminIntegrations = false
	utils.SetDefaultRolesBasedOnConfig()
	_, resp = Client.CreateOAuthApp(oapp)
	CheckNoError(t, resp)

	oapp.Name = ""
	_, resp = AdminClient.CreateOAuthApp(oapp)
	CheckBadRequestStatus(t, resp)

	if r, err := Client.DoApiPost("/oauth/apps", "garbage"); err == nil {
		t.Fatal("should have failed")
	} else {
		if r.StatusCode != http.StatusBadRequest {
			t.Log("actual: " + strconv.Itoa(r.StatusCode))
			t.Log("expected: " + strconv.Itoa(http.StatusBadRequest))
			t.Fatal("wrong status code")
		}
	}

	Client.Logout()
	_, resp = Client.CreateOAuthApp(oapp)
	CheckUnauthorizedStatus(t, resp)

	utils.Cfg.ServiceSettings.EnableOAuthServiceProvider = false
	oapp.Name = GenerateTestAppName()
	_, resp = AdminClient.CreateOAuthApp(oapp)
	CheckNotImplementedStatus(t, resp)
}

func TestGetOAuthApps(t *testing.T) {
	th := Setup().InitBasic().InitSystemAdmin()
	defer TearDown()
	Client := th.Client
	AdminClient := th.SystemAdminClient

	enableOAuth := utils.Cfg.ServiceSettings.EnableOAuthServiceProvider
	adminOnly := *utils.Cfg.ServiceSettings.EnableOnlyAdminIntegrations
	defer func() {
		utils.Cfg.ServiceSettings.EnableOAuthServiceProvider = enableOAuth
		*utils.Cfg.ServiceSettings.EnableOnlyAdminIntegrations = adminOnly
	}()
	utils.Cfg.ServiceSettings.EnableOAuthServiceProvider = true
	*utils.Cfg.ServiceSettings.EnableOnlyAdminIntegrations = false
	utils.SetDefaultRolesBasedOnConfig()

	oapp := &model.OAuthApp{Name: GenerateTestAppName(), Homepage: "https://nowhere.com", Description: "test", CallbackUrls: []string{"https://nowhere.com"}}

	rapp, resp := AdminClient.CreateOAuthApp(oapp)
	CheckNoError(t, resp)

	oapp.Name = GenerateTestAppName()
	rapp2, resp := Client.CreateOAuthApp(oapp)
	CheckNoError(t, resp)

	apps, resp := AdminClient.GetOAuthApps(0, 1000)
	CheckNoError(t, resp)

	found1 := false
	found2 := false
	for _, a := range apps {
		if a.Id == rapp.Id {
			found1 = true
		}
		if a.Id == rapp2.Id {
			found2 = true
		}
	}

	if !found1 || !found2 {
		t.Fatal("missing oauth app")
	}

	apps, resp = AdminClient.GetOAuthApps(1, 1)
	CheckNoError(t, resp)

	if len(apps) != 1 {
		t.Fatal("paging failed")
	}

	apps, resp = Client.GetOAuthApps(0, 1000)
	CheckNoError(t, resp)

	if len(apps) != 1 && apps[0].Id != rapp2.Id {
		t.Fatal("wrong apps returned")
	}

	*utils.Cfg.ServiceSettings.EnableOnlyAdminIntegrations = true
	utils.SetDefaultRolesBasedOnConfig()

	_, resp = Client.GetOAuthApps(0, 1000)
	CheckForbiddenStatus(t, resp)

	Client.Logout()

	_, resp = Client.GetOAuthApps(0, 1000)
	CheckUnauthorizedStatus(t, resp)

	utils.Cfg.ServiceSettings.EnableOAuthServiceProvider = false
	_, resp = AdminClient.GetOAuthApps(0, 1000)
	CheckNotImplementedStatus(t, resp)
}

func TestGetOAuthApp(t *testing.T) {
	th := Setup().InitBasic().InitSystemAdmin()
	defer TearDown()
	Client := th.Client
	AdminClient := th.SystemAdminClient

	enableOAuth := utils.Cfg.ServiceSettings.EnableOAuthServiceProvider
	adminOnly := *utils.Cfg.ServiceSettings.EnableOnlyAdminIntegrations
	defer func() {
		utils.Cfg.ServiceSettings.EnableOAuthServiceProvider = enableOAuth
		*utils.Cfg.ServiceSettings.EnableOnlyAdminIntegrations = adminOnly
	}()
	utils.Cfg.ServiceSettings.EnableOAuthServiceProvider = true
	*utils.Cfg.ServiceSettings.EnableOnlyAdminIntegrations = false
	utils.SetDefaultRolesBasedOnConfig()

	oapp := &model.OAuthApp{Name: GenerateTestAppName(), Homepage: "https://nowhere.com", Description: "test", CallbackUrls: []string{"https://nowhere.com"}}

	rapp, resp := AdminClient.CreateOAuthApp(oapp)
	CheckNoError(t, resp)

	oapp.Name = GenerateTestAppName()
	rapp2, resp := Client.CreateOAuthApp(oapp)
	CheckNoError(t, resp)

	rrapp, resp := AdminClient.GetOAuthApp(rapp.Id)
	CheckNoError(t, resp)

	if rapp.Id != rrapp.Id {
		t.Fatal("wrong app")
	}

	if rrapp.ClientSecret == "" {
		t.Fatal("should not be sanitized")
	}

	rrapp2, resp := AdminClient.GetOAuthApp(rapp2.Id)
	CheckNoError(t, resp)

	if rapp2.Id != rrapp2.Id {
		t.Fatal("wrong app")
	}

	if rrapp2.ClientSecret == "" {
		t.Fatal("should not be sanitized")
	}

	_, resp = Client.GetOAuthApp(rapp2.Id)
	CheckNoError(t, resp)

	_, resp = Client.GetOAuthApp(rapp.Id)
	CheckForbiddenStatus(t, resp)

	*utils.Cfg.ServiceSettings.EnableOnlyAdminIntegrations = true
	utils.SetDefaultRolesBasedOnConfig()

	_, resp = Client.GetOAuthApp(rapp2.Id)
	CheckForbiddenStatus(t, resp)

	Client.Logout()

	_, resp = Client.GetOAuthApp(rapp2.Id)
	CheckUnauthorizedStatus(t, resp)

	_, resp = AdminClient.GetOAuthApp("junk")
	CheckBadRequestStatus(t, resp)

	_, resp = AdminClient.GetOAuthApp(model.NewId())
	CheckNotFoundStatus(t, resp)

	utils.Cfg.ServiceSettings.EnableOAuthServiceProvider = false
	_, resp = AdminClient.GetOAuthApp(rapp.Id)
	CheckNotImplementedStatus(t, resp)
}

func TestGetOAuthAppInfo(t *testing.T) {
	th := Setup().InitBasic().InitSystemAdmin()
	defer TearDown()
	Client := th.Client
	AdminClient := th.SystemAdminClient

	enableOAuth := utils.Cfg.ServiceSettings.EnableOAuthServiceProvider
	adminOnly := *utils.Cfg.ServiceSettings.EnableOnlyAdminIntegrations
	defer func() {
		utils.Cfg.ServiceSettings.EnableOAuthServiceProvider = enableOAuth
		*utils.Cfg.ServiceSettings.EnableOnlyAdminIntegrations = adminOnly
	}()
	utils.Cfg.ServiceSettings.EnableOAuthServiceProvider = true
	*utils.Cfg.ServiceSettings.EnableOnlyAdminIntegrations = false
	utils.SetDefaultRolesBasedOnConfig()

	oapp := &model.OAuthApp{Name: GenerateTestAppName(), Homepage: "https://nowhere.com", Description: "test", CallbackUrls: []string{"https://nowhere.com"}}

	rapp, resp := AdminClient.CreateOAuthApp(oapp)
	CheckNoError(t, resp)

	oapp.Name = GenerateTestAppName()
	rapp2, resp := Client.CreateOAuthApp(oapp)
	CheckNoError(t, resp)

	rrapp, resp := AdminClient.GetOAuthAppInfo(rapp.Id)
	CheckNoError(t, resp)

	if rapp.Id != rrapp.Id {
		t.Fatal("wrong app")
	}

	if rrapp.ClientSecret != "" {
		t.Fatal("should be sanitized")
	}

	rrapp2, resp := AdminClient.GetOAuthAppInfo(rapp2.Id)
	CheckNoError(t, resp)

	if rapp2.Id != rrapp2.Id {
		t.Fatal("wrong app")
	}

	if rrapp2.ClientSecret != "" {
		t.Fatal("should be sanitized")
	}

	_, resp = Client.GetOAuthAppInfo(rapp2.Id)
	CheckNoError(t, resp)

	_, resp = Client.GetOAuthAppInfo(rapp.Id)
	CheckNoError(t, resp)

	*utils.Cfg.ServiceSettings.EnableOnlyAdminIntegrations = true
	utils.SetDefaultRolesBasedOnConfig()

	_, resp = Client.GetOAuthAppInfo(rapp2.Id)
	CheckNoError(t, resp)

	Client.Logout()

	_, resp = Client.GetOAuthAppInfo(rapp2.Id)
	CheckUnauthorizedStatus(t, resp)

	_, resp = AdminClient.GetOAuthAppInfo("junk")
	CheckBadRequestStatus(t, resp)

	_, resp = AdminClient.GetOAuthAppInfo(model.NewId())
	CheckNotFoundStatus(t, resp)

	utils.Cfg.ServiceSettings.EnableOAuthServiceProvider = false
	_, resp = AdminClient.GetOAuthAppInfo(rapp.Id)
	CheckNotImplementedStatus(t, resp)
}