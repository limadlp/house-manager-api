# 📜 Shopping List API with Firestore and WebSockets

This is a Go API using the **Gin** framework and **Firestore** database to create collaborative shopping lists in real-time. The system supports WebSockets to notify users about changes in the lists.

## 📌 Technologies Used

- **Go (Golang)**
- **Gin (HTTP Framework for Go)**
- **Google Firestore (Firebase NoSQL Database)**
- **WebSockets (for real-time communication)**

---

## 🚀 Firebase Configuration

1. **Create a Firebase project** at [Firebase Console](https://console.firebase.google.com/).
2. **Enable Firestore** in "Native" mode.
3. **Generate a service account key**:
   - Go to **Project Settings** → **Service Accounts**.
   - Click on **Generate new private key** and download the JSON file.
4. **Move the JSON file** to `config/firebase_credentials.json`.
5. **Set the environment variable**:

   export GOOGLE_APPLICATION_CREDENTIALS="config/firebase_credentials.json"

6. **Start the API**:

   go run main.go

---

## 🔥 API Endpoints

### 📜 Shopping Lists

#### **Create a new list**

POST /lists

**Body:**

{
"name": "Pharmacy"
}

#### **List all lists**

GET /lists

#### **Get a specific list** (with all items)

GET /lists/{id}

#### **Delete a list**

DELETE /lists/{id}

---

### 🛒 Items within a list

#### **Add an item to a list**

POST /lists/{id}/items

**Body:**

{
"item": "Antiallergic",
"user": "John"
}

#### **Edit an item in the list**

PUT /lists/{id}/items/{index}

**Body:**

{
"checked": true
}

#### **Remove an item from the list**

DELETE /lists/{id}/items/{index}

---

### 📡 WebSocket for Real-Time Updates

- **Connect to WebSocket:**

  ws://localhost:8080/ws

- All changes (creation, editing, and removal of lists/items) are automatically notified.

---

## 🎯 Project Architecture

The project uses a **layered architecture**, organized as follows:

```bash
/project-root
│── config/ # Firestore configuration
│── controllers/ # API controllers (endpoint logic)
│── models/ # Data structures
│── repositories/ # Database interaction
│── routes/ # Route definitions
│── main.go # API initialization
```

### ❌ Why we **DON'T** use the "usecases" layer?

The **usecases** layer is typically used to decouple business rules from the interface (controllers). However, in this API:

1. The business logic is simple and already well organized in **controllers** and **repositories**.
2. Excessive separation would make the project unnecessarily complex.
3. The API is small and focused on REST service, without extensive business rules.

If the project grows, the `usecases` layer can be added later.

---

## 🛠 Testing with `curl` or Postman

### 🔍 Creating a list:

curl -X POST "http://localhost:8080/lists" -H "Content-Type: application/json" -d '{"name": "Pharmacy"}'

### 📌 Listing all lists:

curl -X GET "http://localhost:8080/lists"

### 🎧 Testing WebSocket:

wscat -c ws://localhost:8080/ws

**Or use an online WebSocket client:** [WebSocket Tester](https://www.piesocket.com/websocket-tester)

---

## 📢 Contributing

1. Clone the repository:

   git clone https://github.com/your-username/shopping-list-api.git

2. Create a branch:

   git checkout -b my-feature

3. Make changes and submit a Pull Request!

---

## 📄 License

This API is under the MIT License. Feel free to use and modify it!

🚀 **Now you have everything you need to run and test the API!**
