/**
 * Shared TypeScript interfaces for User domain.
 * These types are used by the Next.js frontend and match the proto definitions.
 */

export interface User {
  id: string;
  email: string;
  displayName: string;
  avatarUrl: string;
  status: UserStatus;
  createdAt: string;
  updatedAt: string;
}

export type UserStatus = "online" | "offline" | "away";

export interface Contact {
  id: string;
  userId: string;
  contactUserId: string;
  nickname: string;
  createdAt: string;
}

export interface UserProfile {
  id: string;
  email: string;
  displayName: string;
  avatarUrl: string;
}
