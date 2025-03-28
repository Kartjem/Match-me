package models

type User struct {
	UserID             int       `json:"user_id"`
	Email              string    `json:"email"`
	Password           string    `json:"password"`
	Fname              *string   `json:"fname"`
	Surname            *string   `json:"surname"`
	Gender             *string   `json:"gender"`
	Birthdate          *string   `json:"birthdate"`
	About              *string   `json:"about"`
	Hobbies            *[]string `json:"hobbies"`
	Interests          *[]string `json:"interests"`
	Country            *string   `json:"country"`
	City               *string   `json:"city"`
	LookingForGender   *string   `json:"looking_for_gender"`
	LookingForMinAge   *int      `json:"looking_for_min_age"`
	LookingForMaxAge   *int      `json:"looking_for_max_age"`
	Picture            *string   `json:"profile_picture_url"`
	PreferredHobbies   *[]string `json:"preferred_hobbies"`
	PreferredInterests *[]string `json:"preferred_interests"`
}

type Chat struct {
	ID         int    `json:"id"`
	SenderID   int    `json:"sender_id"`
	ReceiverID int    `json:"receiver_id"`
	Message    string `json:"message"`
	CreatedAt  string `json:"created_at"`
	Delivered  bool   `json:"delivered"`
}

type Message struct {
	SenderID   int    `json:"sender_id"`
	ReceiverID int    `json:"receiver_id"`
	Message    string `json:"message"`
	CreatedAt  string `json:"createdAt"`
}
