import axios, { AxiosInstance } from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

// API Client Instance
const apiClient: AxiosInstance = axios.create({
  baseURL: `${API_URL}/api`,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Types
export interface Mailbox {
  address: string;
  createdAt: string;
  expiresAt: string;
  expiryTime: number;
}

export interface EmailListItem {
  id: string;
  from: string;
  subject: string;
  receivedAt: string;
  hasAttachments: boolean;
}

export interface Email {
  id: string;
  address: string;
  from: string;
  to: string[];
  subject: string;
  htmlBody: string;
  textBody: string;
  receivedAt: string;
  expiresAt: string;
  size: number;
  hasAttachments: boolean;
}

export interface CreateMailboxResponse {
  mailbox: Mailbox;
}

export interface GetMailboxResponse {
  mailbox: Mailbox;
}

export interface GetEmailsResponse {
  emails: EmailListItem[];
  count: number;
}

export interface GetEmailResponse {
  email: Email;
}

// API Functions

/**
 * Create a new temporary mailbox
 */
export async function createMailbox(expiryMinutes: number): Promise<Mailbox> {
  const response = await apiClient.post<CreateMailboxResponse>('/mailbox', {
    expiryMinutes,
  });
  return response.data.mailbox;
}

/**
 * Get mailbox information
 */
export async function getMailbox(address: string): Promise<Mailbox> {
  const response = await apiClient.get<GetMailboxResponse>(`/mailbox/${address}`);
  return response.data.mailbox;
}

/**
 * Delete a mailbox and all its emails
 */
export async function deleteMailbox(address: string): Promise<void> {
  await apiClient.delete(`/mailbox/${address}`);
}

/**
 * Get all emails for a mailbox
 */
export async function getEmails(address: string): Promise<EmailListItem[]> {
  const response = await apiClient.get<GetEmailsResponse>(`/mailbox/${address}/emails`);
  return response.data.emails || [];
}

/**
 * Get a single email by ID
 */
export async function getEmail(emailId: string): Promise<Email> {
  const response = await apiClient.get<GetEmailResponse>(`/email/${emailId}`);
  return response.data.email;
}

/**
 * Delete a single email
 */
export async function deleteEmail(emailId: string): Promise<void> {
  await apiClient.delete(`/email/${emailId}`);
}

/**
 * Check backend health
 */
export async function checkHealth(): Promise<boolean> {
  try {
    const response = await axios.get(`${API_URL}/health`, { timeout: 5000 });
    return response.status === 200;
  } catch {
    return false;
  }
}

// Error handling helper
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response) {
      // Server responded with error
      const message = error.response.data?.message || error.response.data?.error || 'An error occurred';
      throw new Error(message);
    } else if (error.request) {
      // Request made but no response
      throw new Error('No response from server. Please check your connection.');
    } else {
      // Error setting up request
      throw new Error(error.message || 'An error occurred');
    }
  }
);

export default apiClient;

