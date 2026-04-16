import express from "express";
import { createServer as createViteServer } from "vite";
import path from "path";
import { fileURLToPath } from "url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

interface Notification {
  id: string;
  destination: string;
  channel: "email" | "telegram";
  message: string;
  data_sent_at: string;
  status: "pending" | "sent" | "canceled";
  createdAt: string;
}

let notifications: Notification[] = [];

async function startServer() {
  const app = express();
  const PORT = 3000;

  app.use(express.json());

  // API Routes (matching Go router paths)
  app.get("/all", (req, res) => {
    console.log("GET /all requested");
    res.json({ notes: notifications });
  });

  app.post("/create", (req, res) => {
    console.log("POST /create requested", req.body);
    const { destination, channel, message, data_sent_at } = req.body;
    
    if (!destination || !channel || !message || !data_sent_at) {
      console.warn("Missing fields in request body");
      return res.status(400).json({ error: "Missing required fields" });
    }

    const channelStr = String(channel).toLowerCase();
    if (channelStr !== "email" && channelStr !== "telegram") {
      return res.status(400).json({ error: "Invalid channel. Allowed: email, telegram" });
    }

    const newNote: Notification = {
      id: Math.random().toString(36).substring(2, 9),
      destination: String(destination),
      channel: channelStr as "email" | "telegram",
      message: String(message),
      data_sent_at: String(data_sent_at),
      status: "pending",
      createdAt: new Date().toISOString()
    };

    notifications.push(newNote);
    res.status(201).json(newNote);
  });

  app.delete("/cancel/:id", (req, res) => {
    const { id } = req.params;
    const note = notifications.find(n => n.id === id);
    if (!note) {
      return res.status(404).json({ error: "Notification not found" });
    }
    note.status = "canceled";
    res.json({ success: true });
  });

  // Background worker to "send" notifications
  setInterval(() => {
    const now = new Date();
    notifications.forEach(note => {
      if (note.status === "pending" && new Date(note.data_sent_at) <= now) {
        note.status = "sent";
        console.log(`[Notification Sent] To: ${note.destination} via ${note.channel} | Msg: ${note.message}`);
      }
    });
  }, 10000); // Check every 10 seconds

  // Vite middleware for development
  if (process.env.NODE_ENV !== "production") {
    const vite = await createViteServer({
      server: { middlewareMode: true },
      appType: "spa",
    });
    app.use(vite.middlewares);
  } else {
    const distPath = path.join(process.cwd(), "dist");
    app.use(express.static(distPath));
    app.get("*", (req, res) => {
      res.sendFile(path.join(distPath, "index.html"));
    });
  }

  app.listen(PORT, "0.0.0.0", () => {
    console.log(`Server running on http://app:${PORT}`);
  });
}

startServer();
