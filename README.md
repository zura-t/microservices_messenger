### Functional requirements

1. User registration (by email and password + Oauth)
2. Login/authorization (by email and password + Oauth)
3. Editing user profile (nickname - unique, information about yourself, avatar)
4. Search for users by nickname
5. Add user as friend
6. Remove user from friends
7. Confirm or reject friend request
8. View your friends list (confirmed and unconfirmed)
9. ​​Write a message to a friend
10. Receive a message from a chat with a user

### Non-functional requirements

1. app should work on desktop, and mobile phone

### Rest Api

#### ApiGateway
- ``Post`` _/register_ RegisterUser()
- `Post` */login* Login()
- `Post` */auth/google* AuthorizeGoogle()
- `Post` */auth/google/callback* AuthorizeGoogleCallback()
- `Post` */refresh_token* RefreshToken()
- `Post` _/logout_ Logout() <br/> <br/>

- `Get` */accounts/* GetProfileByUsername()
- `Get` _/accounts/profile_ GetProfile()
- `Post` */accounts/avatar* UploadAvatar()
- `Patch` */accounts/profile* UpdateProfile()
- `Delete` _/accounts/profile_ DeleteProfile() <br/> <br/>

- `Post` */relations* SendFriendRequest()
- `Delete` */relations* RemoveUserFromFriends()
- `Post` */relations/request* ConfirmOrRejectFriendRequest()
- `Get` _/relations_ ViewFriendsList() <br/> <br/>

- `Post` */chat/message* WriteMessage()
- `Get` */chat/messages* GetMessages()

#### Accounts
- ``Post`` _/createUser_ CreateUser()
- `Post` */validateUser* ValidateUser()

- `Get` */accounts/* GetProfileByUsername()
- `Get` _/accounts/profile_ GetProfile()
- `Post` */accounts/avatar* UploadAvatar()
- `Patch` */accounts/profile* UpdateProfile()
- `Delete` _/accounts/profile_ DeleteProfile()

#### Mailer System
- `Post` */sendEmail* SendEmail()

#### Relations
- `Post` */relations* SendFriendRequest()
- `Delete` */relations* RemoveUserFromFriends()
- `Post` */relations/request* ConfirmOrRejectFriendRequest()
- `Get` _/relations_ ViewFriendsList()

#### Chat System
- `Post` */chat/message* WriteMessage()
- `Get` */chat/messages* GetMessages()