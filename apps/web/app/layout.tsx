import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "CyberDitto",
  description: "Attack your twin, not your production.",
};

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
