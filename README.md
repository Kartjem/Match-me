# Match-Me Web

A recommendation application that connects users based on their interests, preferences.

---

## Features

- **User Registration & Authentication**
  - Secure registration with email and password.
  - JWT-based authentication with login and logout.
- **Profile Management**
  - Complete your profile with a minimum of five biographical data points.
  - Upload, change, or remove your profile picture.
  - Specify your location from a predefined list.
- **Matching & Recommendations**
  - Recommendation algorithm using at least five biographical data points.
  - Only shows recommendations when the profile is complete.
  - Displays a maximum of 10 recommendations at a time.
  - Dismiss recommendations that are not interesting.
- **Connections & Chat**
  - Send connection requests and accept or reject incoming requests.
  - Disconnect from users if they are no longer interesting.
  - Real-time chat functionality.
- **Security & Privacy**
  - Private email addresses—only visible to the profile owner.
  - Endpoints enforce proper access control.
- **Admin Tools**
  - Load fictitious users for testing.
  - Reset the database with a simple admin endpoint.

---

## Setup and Installation

### Prerequisites

- [Go](https://golang.org/)
- [Node.js](https://nodejs.org/) and npm
- [PostgreSQL](https://www.postgresql.org/) (make sure it is installed and running)

### Backend

1. **Clone the Repository:**

   ```bash
   git clone https://gitea.kood.tech/artjomkulikovski/match-me.git
   cd match-me/backend
   ```

2. **Configure Database Connection**
    Update the DSN in internal/db/db.go to match your PostgreSQL settings.
3. **Apply Migrations and Run the Server**

    ```bash
    go run server.go
    ```

### Frontend

1. **Navigate to the Frontend Directory:**

      ```bash
      cd match-me/frontend
      ```

2. **Install Dependencies:**

   ```bash
   npm install
   ```

3. **Run the React Application:**

    ```bash
    npm start
    ```

---

## Usage

1. **Register:**  
   Create a new account by navigating to [http://localhost:3000/register](http://localhost:3000/register) and providing your email and password.

2. **Log In:**  
   Log in at [http://localhost:3000/login](http://localhost:3000/login) using your registered credentials.

3. **Complete Your Profile:**  
   After logging in, go to [http://localhost:3000/profile](http://localhost:3000/profile) to complete your profile with all the required details—such as your name, gender, birthdate, location, hobbies, interests, and a profile picture. (This step is mandatory before you can view any recommendations.)

4. **View Recommendations:**  
   Once your profile is complete, navigate to [http://localhost:3000/swipe](http://localhost:3000/swipe) to view your list of recommended connections. You can like (send a connection request) or dismiss a recommendation. Dismissed recommendations won’t be shown again.

5. **Send or Accept Connection Requests:**  
    When you like a recommendation, a connection request is sent automatically. If the other user also likes your profile, the connection is established.

6. **Chat with Connected Users:**  
   After you’re connected with another user, you can chat in real time. Visit [http://localhost:3000/chat](http://localhost:3000/chat) to see your list of chats, receive unread message notifications, and send messages. All chat messages are timestamped.

7. **Update Your Profile:**  
   You can update your profile information at any time by returning to [http://localhost:3000/profile](http://localhost:3000/profile) and editing your details.

8. **Admin Actions:**  
   For demonstration and testing, access the admin panel at [http://localhost:3000/admin](http://localhost:3000/admin) where you can load fake users or reset the database with a single click.
