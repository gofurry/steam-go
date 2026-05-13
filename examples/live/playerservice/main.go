package main

import (
	"fmt"

	"github.com/GoFurry/steam-go/api/playerservice"
	"github.com/GoFurry/steam-go/examples/live/internal/realtest"
)

func main() {
	cfg, err := realtest.LoadConfig()
	if err != nil {
		realtest.Fatalf("load config failed: %v", err)
	}

	client, err := realtest.NewClient(cfg)
	if err != nil {
		realtest.Fatalf("create client failed: %v", err)
	}
	defer client.Close()

	ctx := realtest.BackgroundContext()
	realtest.PrintProxy(cfg)

	if realtest.RequireAPIKey(cfg) {
		fmt.Println("== PlayerService.GetOwnedGames ==")
		ownedGamesResp, err := client.API.PlayerService.GetOwnedGames(
			ctx,
			realtest.DefaultSteamID,
			&playerservice.GetOwnedGamesOptions{IncludePlayedFreeGames: true},
		)
		if err != nil {
			realtest.Fatalf("GetOwnedGames failed: %v", err)
		}
		fmt.Printf("game_count=%d\n", ownedGamesResp.Response.GameCount)
		for i, game := range ownedGamesResp.Response.Games {
			if i >= 10 {
				break
			}
			fmt.Printf("[%d] appid=%d name=%s playtime_forever=%d\n", i+1, game.AppID, game.Name, game.PlaytimeForever)
		}

		fmt.Println("\n== PlayerService.GetAnimatedAvatar ==")
		avatarResp, err := client.API.PlayerService.GetAnimatedAvatar(ctx, realtest.DefaultSteamID, &playerservice.PlayerProfileItemOptions{Language: "zh"})
		if err != nil {
			realtest.Fatalf("GetAnimatedAvatar failed: %v", err)
		}
		fmt.Printf("communityitemid=%s name=%s\n", avatarResp.Response.Avatar.CommunityItemID, avatarResp.Response.Avatar.Name)

		fmt.Println("\n== PlayerService.GetAvatarFrame ==")
		frameResp, err := client.API.PlayerService.GetAvatarFrame(ctx, realtest.DefaultSteamID, &playerservice.PlayerProfileItemOptions{Language: "zh"})
		if err != nil {
			realtest.Fatalf("GetAvatarFrame failed: %v", err)
		}
		fmt.Printf("communityitemid=%s name=%s\n", frameResp.Response.AvatarFrame.CommunityItemID, frameResp.Response.AvatarFrame.Name)

		fmt.Println("\n== PlayerService.GetMiniProfileBackground ==")
		miniProfileResp, err := client.API.PlayerService.GetMiniProfileBackground(ctx, realtest.DefaultSteamID, &playerservice.PlayerProfileItemOptions{Language: "zh"})
		if err != nil {
			realtest.Fatalf("GetMiniProfileBackground failed: %v", err)
		}
		fmt.Printf("communityitemid=%s name=%s\n", miniProfileResp.Response.ProfileBackground.CommunityItemID, miniProfileResp.Response.ProfileBackground.Name)

		fmt.Println("\n== PlayerService.GetProfileBackground ==")
		profileBackgroundResp, err := client.API.PlayerService.GetProfileBackground(ctx, realtest.DefaultSteamID, &playerservice.PlayerProfileItemOptions{Language: "zh"})
		if err != nil {
			realtest.Fatalf("GetProfileBackground failed: %v", err)
		}
		fmt.Printf("communityitemid=%s name=%s\n", profileBackgroundResp.Response.ProfileBackground.CommunityItemID, profileBackgroundResp.Response.ProfileBackground.Name)

		fmt.Println("\n== PlayerService.GetBadges ==")
		badgesResp, err := client.API.PlayerService.GetBadges(ctx, realtest.DefaultSteamID)
		if err != nil {
			realtest.Fatalf("GetBadges failed: %v", err)
		}
		fmt.Printf("badges=%d level=%d xp=%d\n", len(badgesResp.Response.Badges), badgesResp.Response.PlayerLevel, badgesResp.Response.PlayerXP)

		fmt.Println("\n== PlayerService.GetCommunityBadgeProgress ==")
		communityBadgeResp, err := client.API.PlayerService.GetCommunityBadgeProgress(ctx, realtest.DefaultSteamID)
		if err != nil {
			realtest.Fatalf("GetCommunityBadgeProgress failed: %v", err)
		}
		fmt.Printf("quests=%d\n", len(communityBadgeResp.Response.Quests))

		fmt.Println("\n== PlayerService.GetFavoriteBadge ==")
		favoriteBadgeResp, err := client.API.PlayerService.GetFavoriteBadge(ctx, realtest.DefaultSteamID)
		if err != nil {
			realtest.Fatalf("GetFavoriteBadge failed: %v", err)
		}
		fmt.Printf("has_favorite_badge=%t appid=%d level=%d\n", favoriteBadgeResp.Response.HasFavoriteBadge, favoriteBadgeResp.Response.AppID, favoriteBadgeResp.Response.Level)

		fmt.Println("\n== PlayerService.GetPlayerLinkDetails ==")
		linkDetailsResp, err := client.API.PlayerService.GetPlayerLinkDetails(ctx, []string{realtest.DefaultSteamID})
		if err != nil {
			realtest.Fatalf("GetPlayerLinkDetails failed: %v", err)
		}
		fmt.Printf("accounts=%d\n", len(linkDetailsResp.Response.Accounts))

		fmt.Println("\n== PlayerService.GetProfileCustomization ==")
		customizationResp, err := client.API.PlayerService.GetProfileCustomization(
			ctx,
			realtest.DefaultSteamID,
			&playerservice.GetProfileCustomizationOptions{
				IncludeInactiveCustomizations:  true,
				IncludePurchasedCustomizations: true,
			},
		)
		if err != nil {
			realtest.Fatalf("GetProfileCustomization failed: %v", err)
		}
		fmt.Printf("customizations=%d slots_available=%d\n", len(customizationResp.Response.Customizations), customizationResp.Response.SlotsAvailable)

		fmt.Println("\n== PlayerService.GetProfileItemsEquipped ==")
		equippedResp, err := client.API.PlayerService.GetProfileItemsEquipped(ctx, realtest.DefaultSteamID, &playerservice.PlayerProfileItemOptions{Language: "zh"})
		if err != nil {
			realtest.Fatalf("GetProfileItemsEquipped failed: %v", err)
		}
		fmt.Printf("background=%s avatar_frame=%s\n", equippedResp.Response.ProfileBackground.Name, equippedResp.Response.AvatarFrame.Name)

		fmt.Println("\n== PlayerService.GetSteamLevelDistribution ==")
		levelDistributionResp, err := client.API.PlayerService.GetSteamLevelDistribution(ctx, 10)
		if err != nil {
			realtest.Fatalf("GetSteamLevelDistribution failed: %v", err)
		}
		fmt.Printf("player_level_percentile=%.4f\n", levelDistributionResp.Response.PlayerLevelPercentile)

		fmt.Println("\n== PlayerService.GetTopAchievementsForGames ==")
		topAchievementsResp, err := client.API.PlayerService.GetTopAchievementsForGames(
			ctx,
			realtest.DefaultSteamID,
			&playerservice.GetTopAchievementsForGamesOptions{
				Language:        "zh",
				MaxAchievements: 3,
				AppIDs:          []uint32{realtest.DefaultAppID},
			},
		)
		if err != nil {
			realtest.Fatalf("GetTopAchievementsForGames failed: %v", err)
		}
		fmt.Printf("games=%d\n", len(topAchievementsResp.Response.Games))
	}

	if !realtest.RequireAccessToken(cfg) {
		return
	}

	fmt.Println("\n== PlayerService.ClientGetLastPlayedTimes ==")
	lastPlayedResp, err := client.API.PlayerService.ClientGetLastPlayedTimes(
		ctx,
		cfg.AccessToken,
		&playerservice.ClientGetLastPlayedTimesOptions{MinLastPlayed: 100},
	)
	if err != nil {
		realtest.Fatalf("ClientGetLastPlayedTimes failed: %v", err)
	}
	fmt.Printf("games=%d\n", len(lastPlayedResp.Response.Games))

	fmt.Println("\n== PlayerService.GetAchievementsProgress ==")
	includeUnvettedApps := true
	achievementProgressResp, err := client.API.PlayerService.GetAchievementsProgress(
		ctx,
		cfg.AccessToken,
		&playerservice.GetAchievementsProgressOptions{
			SteamID:             realtest.DefaultSteamID,
			Language:            "zh",
			AppIDs:              []uint32{realtest.DefaultAppID, 10},
			IncludeUnvettedApps: &includeUnvettedApps,
		},
	)
	if err != nil {
		realtest.Fatalf("GetAchievementsProgress failed: %v", err)
	}
	fmt.Printf("entries=%d\n", len(achievementProgressResp.Response.AchievementProgress))

	fmt.Println("\n== PlayerService.GetCommunityPreferences ==")
	communityPreferencesResp, err := client.API.PlayerService.GetCommunityPreferences(ctx, cfg.AccessToken)
	if err != nil {
		realtest.Fatalf("GetCommunityPreferences failed: %v", err)
	}
	fmt.Printf("text_filter_setting=%d ignore_friends=%t\n", communityPreferencesResp.Response.Preferences.TextFilterSetting, communityPreferencesResp.Response.Preferences.TextFilterIgnoreFriends)

	fmt.Println("\n== PlayerService.GetNicknameList ==")
	nicknameListResp, err := client.API.PlayerService.GetNicknameList(ctx, cfg.AccessToken)
	if err != nil {
		realtest.Fatalf("GetNicknameList failed: %v", err)
	}
	fmt.Printf("nicknames=%d\n", len(nicknameListResp.Response.Nicknames))

	fmt.Println("\n== PlayerService.GetFriendsGameplayInfo ==")
	friendsGameplayResp, err := client.API.PlayerService.GetFriendsGameplayInfo(ctx, cfg.AccessToken, realtest.DefaultAppID)
	if err != nil {
		realtest.Fatalf("GetFriendsGameplayInfo failed: %v", err)
	}
	fmt.Printf("played_recently=%d played_ever=%d owns=%d wishlist=%d\n",
		len(friendsGameplayResp.Response.PlayedRecently),
		len(friendsGameplayResp.Response.PlayedEver),
		len(friendsGameplayResp.Response.Owns),
		len(friendsGameplayResp.Response.InWishlist),
	)

	fmt.Println("\n== PlayerService.GetProfileItemsOwned ==")
	profileItemsOwnedResp, err := client.API.PlayerService.GetProfileItemsOwned(
		ctx,
		cfg.AccessToken,
		&playerservice.GetProfileItemsOwnedOptions{
			Language: "zh",
			Filters:  []int32{3, 13, 14, 15},
		},
	)
	if err != nil {
		realtest.Fatalf("GetProfileItemsOwned failed: %v", err)
	}
	fmt.Printf("backgrounds=%d mini_backgrounds=%d avatar_frames=%d animated_avatars=%d profile_modifiers=%d\n",
		len(profileItemsOwnedResp.Response.ProfileBackgrounds),
		len(profileItemsOwnedResp.Response.MiniProfileBackgrounds),
		len(profileItemsOwnedResp.Response.AvatarFrames),
		len(profileItemsOwnedResp.Response.AnimatedAvatars),
		len(profileItemsOwnedResp.Response.ProfileModifiers),
	)

	fmt.Println("\n== PlayerService.GetProfileThemesAvailable ==")
	profileThemesResp, err := client.API.PlayerService.GetProfileThemesAvailable(ctx, cfg.AccessToken)
	if err != nil {
		realtest.Fatalf("GetProfileThemesAvailable failed: %v", err)
	}
	fmt.Printf("themes=%d\n", len(profileThemesResp.Response.ProfileThemes))

	fmt.Println("\n== PlayerService.GetPurchasedAndUpgradedProfileCustomizations ==")
	purchasedAndUpgradedResp, err := client.API.PlayerService.GetPurchasedAndUpgradedProfileCustomizations(
		ctx,
		cfg.AccessToken,
		realtest.DefaultSteamID,
	)
	if err != nil {
		realtest.Fatalf("GetPurchasedAndUpgradedProfileCustomizations failed: %v", err)
	}
	fmt.Printf("purchased=%d upgraded=%d\n",
		len(purchasedAndUpgradedResp.Response.PurchasedCustomizations),
		len(purchasedAndUpgradedResp.Response.UpgradedCustomizations),
	)

	fmt.Println("\n== PlayerService.GetRecentlyPlayedGames ==")
	recentlyPlayedResp, err := client.API.PlayerService.GetRecentlyPlayedGames(
		ctx,
		cfg.AccessToken,
		realtest.DefaultSteamID,
		&playerservice.GetRecentlyPlayedGamesOptions{Count: 10},
	)
	if err != nil {
		realtest.Fatalf("GetRecentlyPlayedGames failed: %v", err)
	}
	fmt.Printf("total_count=%d games=%d\n", recentlyPlayedResp.Response.TotalCount, len(recentlyPlayedResp.Response.Games))

	fmt.Println("\n== PlayerService.GetPurchasedProfileCustomizations ==")
	purchasedProfileResp, err := client.API.PlayerService.GetPurchasedProfileCustomizations(ctx, realtest.DefaultSteamID)
	if err != nil {
		realtest.Fatalf("GetPurchasedProfileCustomizations failed: %v", err)
	}
	fmt.Printf("purchased_customizations=%d\n", len(purchasedProfileResp.Response.PurchasedCustomizations))

	fmt.Println("\n== PlayerService.GetSteamLevel ==")
	steamLevelResp, err := client.API.PlayerService.GetSteamLevel(ctx, realtest.DefaultSteamID)
	if err != nil {
		realtest.Fatalf("GetSteamLevel failed: %v", err)
	}
	fmt.Printf("player_level=%d\n", steamLevelResp.Response.PlayerLevel)
}
