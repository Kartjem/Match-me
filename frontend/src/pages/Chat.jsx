import React, { useEffect, useState, useRef, useContext } from "react";
import { useNavigate } from "react-router-dom";
import { AuthContext } from "../context/AuthContext";
import { fetchConnections, fetchUserMinimal } from "../api/api";
import "./Chat.css";

function Chat() {
  const { user, isAuthenticated, loading } = useContext(AuthContext);
  const navigate = useNavigate();
  const [connectedUsers, setConnectedUsers] = useState([]);
  const [activeChat, setActiveChat] = useState(null);
  const [messages, setMessages] = useState([]);
  const [newMessage, setNewMessage] = useState("");
  const [onlineUsers, setOnlineUsers] = useState({});
  const [unreadMessages, setUnreadMessages] = useState({});
  const [chatActivity, setChatActivity] = useState({});
  const [chatLoading, setChatLoading] = useState(true);

  const ws = useRef(null);
  const reconnectAttempt = useRef(0);
  const reconnectTimeout = useRef(null);
  const messagesEndRef = useRef(null);

  const wsUrl = "ws://localhost:8080/ws/chat";
  console.log("Using WebSocket URL:", wsUrl);

  const token = localStorage.getItem("authToken");

  useEffect(() => {
    if (!token || token.split(".").length !== 3) {
      console.error("Invalid or missing auth token. Redirecting to login.");
      navigate("/login");
    }
  }, [token, navigate]);

  const createWebSocket = () => {
    const socket = new WebSocket(wsUrl);

    socket.onopen = () => {
      console.log("WebSocket connected");
      reconnectAttempt.current = 0;
      const connectMsg = {
        type: "connect",
        user_id: String(user.user_id),
        token: token,
      };
      socket.send(JSON.stringify(connectMsg));
    };

    socket.onmessage = (event) => {
      const data = JSON.parse(event.data);
      if (data.type === "message") {
        setMessages((prev) => [...prev, data]);
        const partnerId =
          data.sender_id === user.user_id ? data.receiver_id : data.sender_id;
        setChatActivity((prev) => ({ ...prev, [partnerId]: data.timestamp }));
        setUnreadMessages((prev) => ({
          ...prev,
          [data.sender_id]: (prev[data.sender_id] || 0) + 1,
        }));
      } else if (data.type === "delivered") {
        console.log("Message delivered confirmation received");
      } else if (data.type === "typing") {
        console.log("Typing indicator received:", data);
      }
    };

    socket.onerror = (err) => {
      console.error("WebSocket error:", err);
    };

    socket.onclose = (event) => {
      console.warn("WebSocket closed:", event.reason);
      const delay = Math.min(8000, Math.pow(2, reconnectAttempt.current) * 1000);
      reconnectAttempt.current++;
      reconnectTimeout.current = setTimeout(() => {
        console.log("Attempting to reconnect WebSocket...");
        createWebSocket();
      }, delay);
    };

    ws.current = socket;
  };

  useEffect(() => {
    if (!isAuthenticated || !token || token.split(".").length !== 3) return;
    createWebSocket();
    return () => {
      if (ws.current) ws.current.close();
      if (reconnectTimeout.current) clearTimeout(reconnectTimeout.current);
    };
  }, [isAuthenticated, token]);

  useEffect(() => {
    async function fetchOnlineStatus() {
      try {
        const res = await fetch("http://localhost:8080/users/online-status", {
          credentials: "include",
        });
        if (res.ok) {
          const status = await res.json();
          setOnlineUsers(status);
        }
      } catch (err) {
        console.error("Error fetching online status:", err);
      }
    }
    fetchOnlineStatus();
    const interval = setInterval(fetchOnlineStatus, 10000);
    return () => clearInterval(interval);
  }, []);

  useEffect(() => {
    if (!isAuthenticated) return;
    async function loadData() {
      try {
        const connectionsResponse = await fetchConnections();
        if (connectionsResponse && connectionsResponse.ok) {
          const connectionIds = await connectionsResponse.json();
          const fullUsers = await Promise.all(
            connectionIds.map(async (id) => {
              const res = await fetchUserMinimal(id);
              if (res && res.ok) return res.json();
              return null;
            })
          );
          const validUsers = fullUsers.filter((u) => u !== null);
          setConnectedUsers(validUsers);
          if (!activeChat && validUsers.length > 0) {
            setActiveChat(validUsers[0].id);
          }
        } else {
          setConnectedUsers([]);
        }
      } catch (error) {
        console.error("Error loading chat data:", error);
        setConnectedUsers([]);
      } finally {
        setChatLoading(false);
      }
    }
    loadData();
  }, [isAuthenticated, activeChat]);

  useEffect(() => {
    async function loadChatHistory() {
      if (!activeChat) return;
      try {
        const res = await fetch(`http://localhost:8080/chats?receiver_id=${activeChat}`, {
          method: "GET",
          credentials: "include",
        });
        if (res.ok) {
          const history = await res.json();
          const sortedHistory = history.reverse();
          setMessages(sortedHistory);
          if (sortedHistory.length > 0) {
            const lastTimestamp = sortedHistory[sortedHistory.length - 1].timestamp ||
                                  sortedHistory[sortedHistory.length - 1].createdAt;
            setChatActivity((prev) => ({ ...prev, [activeChat]: lastTimestamp }));
          }
        } else {
          console.error("Failed to load chat history");
        }
      } catch (error) {
        console.error("Error loading chat history:", error);
      }
    }
    loadChatHistory();
  }, [activeChat]);

  useEffect(() => {
    if (activeChat) {
      setUnreadMessages((prev) => {
        const newState = { ...prev };
        delete newState[activeChat];
        return newState;
      });
    }
  }, [activeChat]);

  useEffect(() => {
    if (messagesEndRef.current) {
      messagesEndRef.current.scrollIntoView({ behavior: "smooth" });
    }
  }, [messages]);

  const sortedConnectedUsers = connectedUsers
    .slice()
    .sort((a, b) => {
      const aTime = chatActivity[a.id]
        ? new Date(chatActivity[a.id]).getTime()
        : 0;
      const bTime = chatActivity[b.id]
        ? new Date(chatActivity[b.id]).getTime()
        : 0;
      return bTime - aTime;
    });

  const handleSendMessage = (e) => {
    e.preventDefault();
    if (!newMessage.trim() || !activeChat) return;
    const msgObject = {
      type: "message",
      sender_id: user.user_id,
      receiver_id: activeChat,
      content: newMessage.trim(),
      timestamp: new Date().toISOString(),
    };
    if (ws.current && ws.current.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify(msgObject));
      setMessages((prev) => [...prev, msgObject]);
      setChatActivity((prev) => ({ ...prev, [activeChat]: msgObject.timestamp }));
      setNewMessage("");
    } else {
      console.error("WebSocket not open. Cannot send message.");
    }
  };

  const filteredMessages = messages.filter(
    (msg) =>
      (msg.sender_id === activeChat && msg.receiver_id === user.user_id) ||
      (msg.sender_id === user.user_id && msg.receiver_id === activeChat)
  );

  if (loading || chatLoading) {
    return <div>Loading chat...</div>;
  }

  return (
    <div className="chat-container">
      <div className="chat-list">
        <h3>Your Chats</h3>
        {sortedConnectedUsers.length === 0 ? (
          <p>No chats available.</p>
        ) : (
          sortedConnectedUsers.map((chatUser) => (
            <div
              key={chatUser.id}
              className={`chat-user ${activeChat === chatUser.id ? "active" : ""}`}
              onClick={() => setActiveChat(chatUser.id)}
            >
              <span className={`status-dot ${onlineUsers[chatUser.id] ? "online" : "offline"}`}></span>
              <span className="chat-user-name">{chatUser.name || "Unnamed User"}</span>
              {unreadMessages[chatUser.id] > 0 && (
                <span className="unread-badge">{unreadMessages[chatUser.id]}</span>
              )}
            </div>
          ))
        )}
      </div>
      <div className="chat-window">
        {activeChat ? (
          <>
            <h3>
              Chat with {sortedConnectedUsers.find((u) => u.id === activeChat)?.name || "Unknown User"}
            </h3>
            <div className="chat-messages">
              {filteredMessages.length === 0 ? (
                <p>No messages yet. Start the conversation!</p>
              ) : (
                filteredMessages.map((msg, index) => (
                  <div
                    key={index}
                    className={`chat-message ${msg.sender_id === user.user_id ? "sent" : "received"}`}
                  >
                    <p>{msg.content || msg.message}</p>
                    <small>{new Date(msg.timestamp || msg.created_at).toLocaleString()}</small>


                  </div>
                ))
              )}
              <div ref={messagesEndRef} />
            </div>
            <form className="chat-input" onSubmit={handleSendMessage}>
              <input
                type="text"
                placeholder="Type your message..."
                value={newMessage}
                onChange={(e) => setNewMessage(e.target.value)}
              />
              <button type="submit">Send</button>
            </form>
          </>
        ) : (
          <div className="chat-placeholder">
            <p>Please select a chat from the list to start messaging.</p>
          </div>
        )}
      </div>
    </div>
  );
}

export default Chat;
