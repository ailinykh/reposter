package xcom

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
)

func New(logger *slog.Logger) *XComAPI {
	return &XComAPI{
		httpClient: &http.Client{
			Transport: NewAuthTransport(),
		},
		l: logger,
	}
}

type XComAPI struct {
	httpClient *http.Client
	l          *slog.Logger
}

func (x *XComAPI) Get(ctx context.Context, tweetID string) (*Tweet, error) {
	data, _ := json.Marshal(map[string]any{
		"tweetId":                tweetID,
		"withCommunity":          false,
		"includePromotedContent": false,
		"withVoice":              false,
	})
	variables := url.QueryEscape(string(data))

	data, _ = json.Marshal(map[string]bool{
		"creator_subscriptions_tweet_preview_api_enabled":                         true,
		"premium_content_api_read_enabled":                                        false,
		"communities_web_enable_tweet_community_results_fetch":                    true,
		"c9s_tweet_anatomy_moderator_badge_enabled":                               true,
		"responsive_web_grok_analyze_button_fetch_trends_enabled":                 false,
		"responsive_web_grok_analyze_post_followups_enabled":                      false,
		"responsive_web_jetfuel_frame":                                            true,
		"responsive_web_grok_share_attachment_enabled":                            true,
		"responsive_web_grok_annotations_enabled":                                 false,
		"articles_preview_enabled":                                                true,
		"responsive_web_edit_tweet_api_enabled":                                   true,
		"graphql_is_translatable_rweb_tweet_is_translatable_enabled":              true,
		"view_counts_everywhere_api_enabled":                                      true,
		"longform_notetweets_consumption_enabled":                                 true,
		"responsive_web_twitter_article_tweet_consumption_enabled":                true,
		"tweet_awards_web_tipping_enabled":                                        false,
		"responsive_web_grok_show_grok_translated_post":                           false,
		"responsive_web_grok_analysis_button_from_backend":                        true,
		"post_ctas_fetch_enabled":                                                 true,
		"creator_subscriptions_quote_tweet_preview_enabled":                       false,
		"freedom_of_speech_not_reach_fetch_enabled":                               true,
		"standardized_nudges_misinfo":                                             true,
		"tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled": true,
		"longform_notetweets_rich_text_read_enabled":                              true,
		"longform_notetweets_inline_media_enabled":                                true,
		"profile_label_improvements_pcf_label_in_post_enabled":                    true,
		"responsive_web_profile_redirect_enabled":                                 false,
		"rweb_tipjar_consumption_enabled":                                         true,
		"verified_phone_label_enabled":                                            false,
		"responsive_web_grok_image_annotation_enabled":                            true,
		"responsive_web_grok_imagine_annotation_enabled":                          true,
		"responsive_web_grok_community_note_auto_translation_is_enabled":          false,
		"responsive_web_graphql_skip_user_profile_image_extensions_enabled":       false,
		"responsive_web_graphql_timeline_navigation_enabled":                      true,
		"responsive_web_enhance_cards_enabled":                                    false,
	})
	features := url.QueryEscape(string(data))

	urlString := fmt.Sprintf("https://api.x.com/graphql/YTLCpNxePO-aAmb57DAblw/TweetResultByRestId?variables=%s&features=%s", variables, features)
	req, err := http.NewRequestWithContext(ctx, "GET", urlString, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	res, err := x.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}
	defer res.Body.Close()

	var response struct {
		Data struct {
			TweetResult struct {
				Result *Tweet `json:"result"`
			} `json:"tweetResult"`
		} `json:"data"`
	}
	if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode json: %w", err)
	}

	tweet := response.Data.TweetResult.Result

	if tweet == nil {
		return nil, &Error{TypeName: "Hmm... this page doesn’t exist."}
	}

	if tweet.TypeName != "Tweet" {
		return nil, &Error{TypeName: tweet.TypeName, Reason: tweet.Reason}
	}

	return tweet, nil
}
