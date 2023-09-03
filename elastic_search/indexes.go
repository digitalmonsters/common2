package elastic_search

type Index string

// Event Publisher Follow Indexes
const FollowingIndex Index = "following"
const FollowerIndex Index = "follower"
const UserFollowingCounter Index = "user_following_counter"
const UserFollowersCounter Index = "user_followers_counter"
const UserTodayFollowersCounter Index = "user_today_followers_counter"
const UserUniqueFollowersPerDayCount Index = "user_unique_followers_per_day_count"

// Event Publisher Like Indexes
const CounterContentLikes Index = "counter_content_likes"
const CounterLikesToUser Index = "counter_likes_to_user"
const Dislikes Index = "dislikes"
const GeneralUserLikes Index = "general_user_likes"
const CounterUserLikes Index = "counter_user_likes"
const CounterContentHashtag Index = "counter_content_hashtag"
const CounterContentCategory Index = "counter_content_category"

// Event Publisher Spot Indexes
// counter_music_dislikes_to_user
// counter_spot_likes
// counter_user_music_likes
// general_spot_reactions
// counter_music_likes_to_user
// counter_spot_likes_to_user
// counter_user_music_loves
// counter_music_loves_to_user
// counter_spot_loves
// counter_user_spot_dislikes
// counter_spot_dislikes
// counter_spot_loves_to_user
// counter_user_spot_likes
// counter_spot_dislikes_to_user
// counter_user_music_dislikes
// counter_user_spot_loves
const SpotLikesCounter Index = "spot_likes_counter"
const SpotMusicDislikesToUserCounter Index = "spot_music_dislikes_to_user_counter"
const SpotUserMusicLikesCounter Index = "spot_user_music_likes_counter"
const SpotGeneralReactions Index = "spot_general_reactions"
const SpotMusicLikesToUserCounter Index = "spot_music_likes_to_user_counter"
const SpotLikesToUserCounter Index = "spot_likes_to_user_counter"
const SpotUserMusicLovesCounter Index = "spot_user_music_loves_counter"
const SpotMusicLovesToUserCounter Index = "spot_music_loves_to_user_counter"
const SpotLovesCounter Index = "spot_loves_counter"
const SpotUserSpotDislikesCounter Index = "spot_user_spot_dislikes_counter"
const SpotDislikesCounter Index = "spot_dislikes_counter"
const SpotLovesToUserCounter Index = "spot_loves_to_user_counter"
const SpotUserLikesCounter Index = "spot_user_likes_counter"
const SpotDislikesToUserCounter Index = "spot_dislikes_to_user_counter"
const SpotUserMusicDislikesCounter Index = "spot_user_music_dislikes_counter"
const SpotUserSpotLovesCounter Index = "spot_user_spot_loves_counter"

// Event Publisher Category Indexes
// user_category_counter
// mapping_user_categories
// user_category_subscriptions_counter
const UserCategoryCounter Index = "user_category_counter"
const MappingUserCategories Index = "mapping_user_categories"
const UserCategorySubscriptionsCounter Index = "user_category_subscriptions_counter"

// Event Publisher Hashtag Indexes
// user_hashtags_counter
// subscriptions_counter
//
//	subscriptions
const UserHashtagsCounter Index = "user_hashtags_counter"
const SubscriptionsCounter Index = "subscriptions_counter"
const Subscriptions Index = "subscriptions"

// Event Publisher Views Indexes
// counter_category
// counter_user_views_per_day_per_country
// counter_content
// counter_user_views_per_day_per_country_spots
// counter_content_owner
// counter_user_views_per_day_spots
// counter_content_view_per_day_by_user
// counter_user_watch_time
// counter_hashtag
// counter_viewer
// counter_user_view
// counter_user_views_per_day
// user_content

const ViewsCategoryCounter Index = "views_category_counter"
const ViewsUserViewsPerDayPerCountryCounter Index = "views_user_views_per_day_per_country_counter"
const ViewsContentCounter Index = "views_content_counter"
const ViewsUserViewsPerDayPerCountrySpotsCounter Index = "views_user_views_per_day_per_country_spots_counter"
const ViewsContentOwnerCounter Index = "views_content_owner_counter"
const ViewsUserViewsPerDaySpotsCounter Index = "views_user_views_per_day_spots_counter"
const ViewsContentViewPerDayByUserCounter Index = "views_content_view_per_day_by_user_counter"
const ViewsUserWatchTimeCounter Index = "views_user_watch_time_counter"
const ViewsHashtagCounter Index = "views_hashtag_counter"
const ViewsViewerCounter Index = "views_viewer_counter"
const ViewsUserViewCounter Index = "views_user_view_counter"
const ViewsUserViewsPerDayCounter Index = "views_user_views_per_day_counter"
const ViewsUserContentCounter Index = "views_user_content_counter"
