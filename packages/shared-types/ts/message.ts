/**
 * Shared TypeScript interfaces for Message domain.
 */

export type MessageType = "text" | "ai_response" | "system";

export type AIModel = "chatgpt" | "claude" | "gemini";

export interface Message {
  id: string;
  chatId: string;
  senderId: string;
  content: string;
  type: MessageType;
  aiModel?: AIModel;
  createdAt: string;
  updatedAt: string;
  isEdited: boolean;
  readBy: string[];
}

export interface SendMessageRequest {
  chatId: string;
  content: string;
}

export interface MessagePage {
  messages: Message[];
  nextCursor?: string;
  hasMore: boolean;
}
