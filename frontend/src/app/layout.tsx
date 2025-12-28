import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import './globals.css'

const inter = Inter({ subsets: ['latin'] })

export const metadata: Metadata = {
  title: 'Temp Mail 365 - Temporary Email Service',
  description: 'Secure temporary email service. Protect your privacy with disposable email addresses. No registration required.',
  keywords: 'temporary email, disposable email, temp mail, anonymous email, privacy',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body className={inter.className}>{children}</body>
    </html>
  )
}

