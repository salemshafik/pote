/**
 * Shared TypeScript interfaces for WebSocket / Realtime events.
 */

// ---------- Client → Server events ----------

export type ClientEvent =
  | { type: "message:send"; payload: SendMessagePayload }
  | { type: "typing:start"; payload: TypingPayload }
  | { type: "typing:stop"; payload: TypingPayload }
  | { type: "message:read"; payload: ReadReceiptPayload }
  | { type: "presence:update"; payload: PresencePayload };

export interface SendMessagePayload {
  chatId: string;
  content: string;
  tempId: string; // Client-side temporary ID for optimistic UI
}

export interface TypingPayload {
  chatId: string;
}

export interface ReadReceiptPayload {
  chatId: string;
  messageId: string;
}

export interface PresencePayload {
  status: "online" | "away";
}

// ---------- Server → Client events ----------

export type ServerEvent =
  | { type: "message:new"; payload: NewMessagePayload }
  | { type: "message:ack"; payload: MessageAckPayload }
  | { type: "typing:indicator"; payload: TypingIndicatorPayload }
  | { type: "message:read_receipt"; payload: ReadReceiptServerPayload }
  | { type: "presence:changed"; payload: PresenceChangedPayload }
  | { type: "ai:joined"; payload: AIJoinedPayload }
  | { type: "ai:left"; payload: AILeftPayload }
  | { type: "error"; payload: ErrorPayload };

export interface NewMessagePayload {
  id: string;
  chatId: string;
  senderId: string;
  senderName: string;
  content: string;
  type: "text" | "ai_response" | "system";
  aiModel?: string;
  createdAt: string;
}

export interface MessageAckPayload {
  tempId: string;
  messageId: string;
  createdAt: string;
}

export interface TypingIndicatorPayload {
  chatId: string;
  userId: string;
  displayName: string;
  isTyping: boolean;
}

export interface ReadReceiptServerPayload {
  chatId: string;
  messageId: string;
  userId: string;
  readAt: string;
}

export interface PresenceChangedPayload {
  userId: string;
  status: "online" | "offline" | "away";
  lastSeen?: string;
}

export interface AIJoinedPayload {
  chatId: string;
  model: string;
}

export interface AILeftPayload {
  chatId: string;
  model: string;
}

export interface ErrorPayload {
  code: string;
  message: string;
}
