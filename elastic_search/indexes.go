package elastic_search

type Index string

// Event Publisher Follow Indexes
const FollowingIndex Index = "following"
const FollowerIndex Index = "follower"
const UserFollowingCounter Index = "user_following_counter"
const UserFollowerCounter Index = "user_follower_counter"
const UserTodayFollowersCounter Index = "user_today_followers_counter"

// Event Publisher Like Indexes
const CounterContentLikes Index = "counter_content_likes"
const CounterLikesToUser Index = "counter_likes_to_user"
const Dislikes Index = "dislikes"
const GeneralUserLikes Index = "general_user_likes"
const CounterUserLikes Index = "counter_user_likes"
const CounterContentHashtag Index = "counter_content_hashtag"
const CounterContentCategory Index = "counter_content_category"
