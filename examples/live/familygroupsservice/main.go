package main

import (
	"fmt"

	"github.com/GoFurry/steam-go/api/familygroupsservice"
	"github.com/GoFurry/steam-go/examples/live/internal/realtest"
)

func main() {
	cfg, err := realtest.LoadConfig()
	if err != nil {
		realtest.Fatalf("load config failed: %v", err)
	}
	if !realtest.RequireAccessToken(cfg) || !realtest.RequireFamilyGroupID(cfg) {
		return
	}

	client, err := realtest.NewClient(cfg)
	if err != nil {
		realtest.Fatalf("create client failed: %v", err)
	}
	defer client.Close()

	ctx := realtest.BackgroundContext()
	realtest.PrintProxy(cfg)

	fmt.Println("== FamilyGroupsService.GetChangeLog ==")
	changeLogResp, err := client.API.FamilyGroupsService.GetChangeLog(ctx, cfg.FamilyGroupID)
	if err != nil {
		realtest.Fatalf("GetChangeLog failed: %v", err)
	}
	fmt.Printf("changes=%d\n", len(changeLogResp.Response.Changes))

	fmt.Println("\n== FamilyGroupsService.GetFamilyGroup ==")
	familyGroupResp, err := client.API.FamilyGroupsService.GetFamilyGroup(ctx, cfg.FamilyGroupID)
	if err != nil {
		realtest.Fatalf("GetFamilyGroup failed: %v", err)
	}
	fmt.Printf("name=%s members=%d\n", familyGroupResp.Response.Name, len(familyGroupResp.Response.Members))

	fmt.Println("\n== FamilyGroupsService.GetFamilyGroupForUser ==")
	familyForUserResp, err := client.API.FamilyGroupsService.GetFamilyGroupForUser(
		ctx,
		cfg.FamilyGroupID,
		&familygroupsservice.GetFamilyGroupForUserOptions{IncludeFamilyGroupResponse: true},
	)
	if err != nil {
		realtest.Fatalf("GetFamilyGroupForUser failed: %v", err)
	}
	fmt.Printf("family_groupid=%s role=%d\n", familyForUserResp.Response.FamilyGroupID, familyForUserResp.Response.Role)

	fmt.Println("\n== FamilyGroupsService.GetPlaytimeSummary ==")
	playtimeResp, err := client.API.FamilyGroupsService.GetPlaytimeSummary(ctx, cfg.FamilyGroupID)
	if err != nil {
		realtest.Fatalf("GetPlaytimeSummary failed: %v", err)
	}
	fmt.Printf("entries=%d\n", len(playtimeResp.Response.Entries))

	fmt.Println("\n== FamilyGroupsService.GetSharedLibraryApps ==")
	sharedAppsResp, err := client.API.FamilyGroupsService.GetSharedLibraryApps(ctx, cfg.FamilyGroupID)
	if err != nil {
		realtest.Fatalf("GetSharedLibraryApps failed: %v", err)
	}
	fmt.Printf("owner=%s apps=%d\n", sharedAppsResp.Response.OwnerSteamID, len(sharedAppsResp.Response.Apps))
}
