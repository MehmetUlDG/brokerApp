import type { Metadata } from "next";
import { Inter, JetBrains_Mono } from "next/font/google";
import { Toaster } from "sonner";
import "./globals.css";

const inter = Inter({ subsets: ["latin"], variable: '--font-inter' });
const jetbrains = JetBrains_Mono({ subsets: ["latin"], variable: '--font-mono' });

export const metadata: Metadata = {
  title: "TradeOff | Geleceğin Kripto Borsası",
  description: "Güvenilir, hızlı ve düşük komisyonlu kripto para borsası.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="tr" suppressHydrationWarning>
      <body className={`${inter.variable} ${jetbrains.variable} antialiased bg-[var(--bg-primary)] text-[var(--text-primary)] transition-colors duration-300`}>
        {children}
        <Toaster position="bottom-right" />
      </body>
    </html>
  );
}
