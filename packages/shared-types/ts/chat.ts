/**
 * Shared TypeScript interfaces for Chat domain.
 */

export type ChatType = "direct" | "group";

export type MemberRole = "member" | "admin" | "owner";

export interface Chat {
  id: string;
  type: ChatType;
  name: string;
  description: string;
  avatarUrl: string;
  createdBy: string;
  createdAt: string;
  updatedAt: string;
  members: ChatMember[];
}

export interface ChatMember {
  userId: string;
  role: MemberRole;
  joinedAt: string;
}

export interface CreateChatRequest {
  type: ChatType;
  name?: string;
  description?: string;
  memberIds: string[];
}

export interface ChatListItem {
  id: string;
  type: ChatType;
  name: string;
  avatarUrl: string;
  lastMessage?: LastMessage;
  unreadCount: number;
}

export interface LastMessage {
  content: string;
  senderName: string;
  sentAt: string;
}
