# TradeOff — Frontend Implementation Plan

## 1. Proje Özeti

**TradeOff**, Go (Golang) ile yazılmış 4 microservice'ten oluşan bir borsa uygulamasıdır. Bu plan, backend'deki tüm gereksinimleri eksiksiz karşılayan, **FinanceFlow Figma tasarımına** sadık, responsive ve production-ready bir frontend oluşturulmasını kapsar.

> [!IMPORTANT]
> Frontend, `c:\Users\mehme\brokerApp\frontend` dizininde (şu an boş) **Next.js 14 (App Router) + TypeScript + Tailwind CSS v3** ile sıfırdan oluşturulacaktır.

---

## 2. Backend Analizi — API Kontratları

### 2.1 Broker Backend (REST — Port 3000)

| Endpoint | Method | Auth | Request Body | Response |
|---|---|---|---|---|
| `/health` | GET | ❌ | — | `{ status, db, version }` |
| `/api/auth/register` | POST | ❌ | `{ email, password, first_name, last_name }` | `{ token, user: { id, email, first_name, last_name } }` |
| `/api/auth/login` | POST | ❌ | `{ email, password }` | `{ token, user: { id, email, first_name, last_name } }` |
| `/api/wallet` | GET | ✅ JWT | — | `{ id, user_id, balance, btc_balance, updated_at }` |
| `/api/wallet/deposit` | POST | ✅ JWT | `{ amount: "string" }` | Updated Wallet object |
| `/api/wallet/withdraw` | POST | ✅ JWT | `{ amount: "string" }` | Updated Wallet object |
| `/api/orders` | POST | ✅ JWT | `{ symbol, side, type, quantity, price }` | Created Order object |

### 2.2 Payment Service (gRPC → Port 50051 + HTTP Webhook → Port 8081)

| RPC | Params | Response |
|---|---|---|
| `Deposit` | `user_id, amount, currency, stripe_payment_method_id` | `{ transaction_id, status }` |
| `Withdraw` | `user_id, amount, currency, stripe_account_id` | `{ transaction_id, status }` |
| `GetHistory` | `user_id, limit, offset` | `{ transactions[] }` |
| `GetBalance` | `user_id` | `{ usd_balance, btc_balance }` |

> [!NOTE]
> Frontend, gRPC'ye direkt bağlanmaz. Broker Backend üzerinden REST proxy kullanır. Eğer backend'e proxy endpoint eklenemezse, frontend'de gRPC-web gateway veya mock data kullanılır.

### 2.3 Ingestion Service (WebSocket → Kafka)

- Binance `wss://stream.binance.com:443/ws/btcusdt@trade` üzerinden real-time BTCUSDT fiyat verisi çeker
- Kafka `live-prices` topic'ine yazar
- **Frontend Stratejisi:** Frontend, doğrudan Binance WebSocket'e bağlanarak anlık fiyat verisi alır (ingestion-service frontend'e endpoint sunmuyor)

### 2.4 Matching Engine (Kafka Consumer/Producer)

- `new-orders` topic'inden PENDING emirleri okur
- `trade-executed` topic'ine eşleşen işlemleri yazar
- Frontend'e doğrudan API sunmaz — emir durumu broker-backend üzerinden izlenir

### 2.5 Domain Modelleri

```typescript
// User
interface User {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
}

// Wallet
interface Wallet {
  id: string;
  user_id: string;
  balance: string;      // DECIMAL → string
  btc_balance: string;  // DECIMAL → string
  updated_at: string;
}

// Order
interface Order {
  id: string;
  user_id: string;
  symbol: string;       // "BTCUSDT"
  side: "BUY" | "SELL";
  type: "MARKET" | "LIMIT";
  quantity: string;
  price: string;
  status: "PENDING" | "COMPLETED" | "FAILED" | "CANCELED";
  created_at: string;
  updated_at: string;
}

// Transaction (Payment Service)
interface Transaction {
  id: string;
  user_id: string;
  type: "DEPOSIT" | "WITHDRAWAL" | "TRANSFER" | "REFUND";
  amount: string;
  currency: string;
  status: "PENDING" | "COMPLETED" | "FAILED";
  stripe_ref: string;
  created_at: string;
}

// Live Price (Binance WebSocket)
interface LivePrice {
  symbol: string;
  price: string;
  timestamp: number;
}
```

---

## 3. Tasarım Sistemi — FinanceFlow Referansı

### 3.1 Renk Paleti

| Token | Light Mode | Dark Mode | Kullanım |
|---|---|---|---|
| `--bg-primary` | `#FFFFFF` | `#0B0F1A` | Ana arka plan |
| `--bg-secondary` | `#F7F8FA` | `#111827` | Kart / Section arka plan |
| `--bg-tertiary` | `#F0F2F5` | `#1A2035` | Hover / Alt bileşenler |
| `--surface` | `#FFFFFF` | `#1E2640` | Kart yüzeyi |
| `--border` | `#E5E7EB` | `#2A3451` | Kenarlıklar |
| `--text-primary` | `#111827` | `#F9FAFB` | Ana metin |
| `--text-secondary` | `#6B7280` | `#9CA3AF` | İkincil metin |
| `--text-muted` | `#9CA3AF` | `#6B7280` | Muted metin |
| `--accent-primary` | `#2563EB` | `#3B82F6` | CTA / Primary action |
| `--accent-hover` | `#1D4ED8` | `#60A5FA` | CTA hover |
| `--success` | `#10B981` | `#34D399` | Pozitif / BUY / Profit |
| `--danger` | `#EF4444` | `#F87171` | Negatif / SELL / Loss |
| `--warning` | `#F59E0B` | `#FBBF24` | Uyarılar |

### 3.2 Tipografi

| Element | Font | Weight | Size |
|---|---|---|---|
| Display | Inter | 700 | 48-64px |
| H1 | Inter | 700 | 36-48px |
| H2 | Inter | 600 | 24-30px |
| H3 | Inter | 600 | 20-24px |
| Body | Inter | 400 | 16px |
| Body Small | Inter | 400 | 14px |
| Caption | Inter | 500 | 12px |
| Mono (Prices) | JetBrains Mono | 500 | 14-24px |

### 3.3 Spacing & Radius

- **Spacing:** 4px grid sistemi (4, 8, 12, 16, 20, 24, 32, 40, 48, 64)
- **Border Radius:** `sm: 6px`, `md: 8px`, `lg: 12px`, `xl: 16px`, `2xl: 20px`, `full: 9999px`
- **Shadow:** Subtle elevation system — `sm`, `md`, `lg`

### 3.4 Breakpoints

| Breakpoint | Değer | Hedef |
|---|---|---|
| `sm` | 640px | Mobil landscape |
| `md` | 768px | Tablet |
| `lg` | 1024px | Tablet landscape / küçük laptop |
| `xl` | 1280px | Desktop |
| `2xl` | 1536px | Geniş ekran |

---

## 4. Sayfa Yapısı & Routing

```
/                       → Landing Page (Public)
/login                  → Login Page (Public)
/register               → Register Page (Public)
/dashboard              → Dashboard / Overview (Protected)
/trade                  → Trading Page (Protected)
/wallet                 → Wallet & Transactions (Protected)
/settings               → User Settings (Protected)
/404                    → Not Found Page
```

### 4.1 Landing Page (`/`)
- Hero section — "Geleceğin Borsası" tarzı başlık, CTA butonları
- Features bölümü — 3-4 özellik kartı (gerçek zamanlı veri, güvenli işlem, düşük komisyon, vb.)
- Stats bölümü — animasyonlu sayaçlar
- CTA bölümü — Kayıt yönlendirmesi
- Footer — sosyal medya linkleri, copyright

### 4.2 Auth Pages (`/login`, `/register`)
- Sol taraf: form alanı
- Sağ taraf: dekoratif görsel / gradient (desktop)
- Mobilde tek kolon
- Form validasyonu (client-side)
- Error/Success toast mesajları

### 4.3 Dashboard (`/dashboard`)
- Portfolio özeti kartı (toplam USD, BTC değeri)
- Canlı BTC fiyat kartı (WebSocket)
- Mini fiyat grafiği (TradingView Lightweight Charts)
- Hızlı işlem widget'ı
- Son emirler tablosu

### 4.4 Trade Page (`/trade`)
- Sol: Canlı fiyat grafiği (TradingView Lightweight Charts — tam genişlik)
- Sağ: Order form paneli (BUY/SELL toggle, MARKET/LIMIT seçimi, miktar, fiyat)
- Alt: Açık emirler / Emir geçmişi tablosu
- Canlı fiyat ticker bar

### 4.5 Wallet Page (`/wallet`)
- Bakiye özeti kartları (USD + BTC)
- Deposit / Withdraw formları (modal veya inline)
- İşlem geçmişi tablosu (pagination destekli)
- Son aktiviteler timeline

### 4.6 Settings Page (`/settings`)
- Profil bilgileri (isim, e-posta — read-only)
- Tema değiştirme (Light / Dark)
- Dil seçimi (opsiyonel)

---

## 5. Component Hiyerarşisi

```
src/
├── app/                          # Next.js App Router
│   ├── layout.tsx                # Root layout (providers, fonts)
│   ├── page.tsx                  # Landing page
│   ├── (auth)/                   # Auth layout group
│   │   ├── layout.tsx            # Auth layout (split screen)
│   │   ├── login/page.tsx
│   │   └── register/page.tsx
│   ├── (dashboard)/              # Protected layout group
│   │   ├── layout.tsx            # Dashboard layout (sidebar + topbar)
│   │   ├── dashboard/page.tsx
│   │   ├── trade/page.tsx
│   │   ├── wallet/page.tsx
│   │   └── settings/page.tsx
│   ├── not-found.tsx             # 404 page
│   └── globals.css               # Global styles + CSS custom properties
│
├── components/
│   ├── ui/                       # Primitif UI bileşenleri
│   │   ├── Button.tsx
│   │   ├── Input.tsx
│   │   ├── Card.tsx
│   │   ├── Modal.tsx
│   │   ├── Badge.tsx
│   │   ├── Spinner.tsx
│   │   ├── Toast.tsx
│   │   ├── Table.tsx
│   │   ├── Tabs.tsx
│   │   ├── Toggle.tsx
│   │   ├── Skeleton.tsx
│   │   └── ThemeToggle.tsx
│   │
│   ├── layout/                   # Layout bileşenleri
│   │   ├── Navbar.tsx            # Public navbar
│   │   ├── Sidebar.tsx           # Dashboard sidebar
│   │   ├── Topbar.tsx            # Dashboard topbar
│   │   ├── Footer.tsx            # Public footer
│   │   └── MobileMenu.tsx        # Mobile hamburger menu
│   │
│   ├── landing/                  # Landing page bileşenleri
│   │   ├── HeroSection.tsx
│   │   ├── FeatureCard.tsx
│   │   ├── StatsSection.tsx
│   │   └── CTASection.tsx
│   │
│   ├── auth/                     # Auth bileşenleri
│   │   ├── LoginForm.tsx
│   │   └── RegisterForm.tsx
│   │
│   ├── dashboard/                # Dashboard bileşenleri
│   │   ├── PortfolioCard.tsx
│   │   ├── PriceCard.tsx
│   │   ├── QuickTradeWidget.tsx
│   │   ├── RecentOrdersTable.tsx
│   │   └── MiniChart.tsx
│   │
│   ├── trade/                    # Trade page bileşenleri
│   │   ├── PriceChart.tsx        # TradingView Lightweight Charts
│   │   ├── OrderForm.tsx
│   │   ├── OrderBook.tsx         # (Mock / görsel)
│   │   ├── TickerBar.tsx
│   │   └── OrderHistory.tsx
│   │
│   └── wallet/                   # Wallet page bileşenleri
│       ├── BalanceCards.tsx
│       ├── DepositForm.tsx
│       ├── WithdrawForm.tsx
│       ├── TransactionHistory.tsx
│       └── ActivityTimeline.tsx
│
├── lib/                          # Utility katmanı
│   ├── api/
│   │   ├── client.ts             # Axios instance + interceptors
│   │   ├── auth.ts               # register, login
│   │   ├── wallet.ts             # getWallet, deposit, withdraw
│   │   └── orders.ts             # placeOrder
│   ├── ws/
│   │   └── binance.ts            # Binance WebSocket bağlantısı
│   ├── utils/
│   │   ├── formatters.ts         # Para, tarih formatlama
│   │   ├── validators.ts         # Form validasyon fonksiyonları
│   │   └── cn.ts                 # clsx + tailwind-merge utility
│   └── constants.ts              # API URLs, WebSocket URLs
│
├── hooks/                        # Custom React hooks
│   ├── useAuth.ts
│   ├── useWallet.ts
│   ├── useLivePrice.ts           # Binance WS hook
│   ├── useOrders.ts
│   └── useTheme.ts
│
├── stores/                       # Zustand state management
│   ├── authStore.ts              # User + JWT token
│   ├── walletStore.ts            # Wallet state
│   ├── tradeStore.ts             # Orders + live price
│   └── themeStore.ts             # Light/Dark theme
│
├── types/                        # TypeScript type definitions
│   ├── api.ts                    # API response types
│   ├── domain.ts                 # User, Wallet, Order, Transaction
│   └── common.ts                 # Shared utility types
│
└── middleware.ts                  # Next.js middleware (route protection)
```

---

## 6. Teknoloji Yığını & Paketler

| Kategori | Paket | Amaç |
|---|---|---|
| **Framework** | Next.js 14 (App Router) | SSR + CSR + routing |
| **Language** | TypeScript 5 | Tip güvenliği |
| **Styling** | Tailwind CSS v3 | Utility-first CSS |
| **State** | Zustand | Hafif global state |
| **HTTP** | Axios | API istekleri + interceptors |
| **Charts** | lightweight-charts (TradingView) | Finansal grafikler |
| **Icons** | lucide-react | Modern SVG icon kütüphanesi |
| **Animations** | framer-motion | Smooth micro-animations |
| **Forms** | react-hook-form + zod | Form yönetimi + validasyon |
| **Toast** | sonner | Toast notification |
| **Fonts** | @next/font (Inter, JetBrains Mono) | Optimize edilmiş web fontları |
| **Utils** | clsx + tailwind-merge | Conditional className |

---

## 7. Kritik Entegrasyon Detayları

### 7.1 JWT Auth Flow
1. Login/Register → backend'den `{ token, user }` alınır
2. Token `localStorage` + Zustand authStore'da tutulur
3. Axios interceptor: her istekte `Authorization: Bearer <token>` eklenir
4. 401 yanıtı → authStore.logout() → `/login` redirect
5. Next.js middleware: korumalı route'larda token kontrolü

### 7.2 WebSocket — Canlı Fiyat
```
wss://stream.binance.com:443/ws/btcusdt@trade
```
- `useLivePrice` hook ile yönetilir
- Auto-reconnect (exponential backoff)
- Son fiyat `tradeStore`'da tutulur
- PriceChart ve TickerBar bileşenleri subscribe olur

### 7.3 Order Flow
1. OrderForm → `POST /api/orders` (LIMIT veya MARKET)
2. Backend → Kafka → Matching Engine → `trade-executed`
3. Frontend: Emir gönderildikten sonra `PENDING` olarak gösterilir
4. Polling ile emir durumu güncellenir (her 5 saniye)

### 7.4 CORS
Backend zaten CORS middleware içeriyor — `Access-Control-Allow-Origin: *` destekli.

---

## 8. Open Questions

> [!IMPORTANT]
> **gRPC Proxy:** Payment Service (GetHistory, GetBalance) gRPC endpoint'lerine frontend doğrudan erişemez. Broker Backend'e REST proxy endpoint'leri (`GET /api/transactions`, `GET /api/balance`) eklenmeli mi, yoksa frontend'de mock veri ile ilerlenecek mi?

> [!IMPORTANT]
> **WebSocket Gateway:** Ingestion service Kafka'ya yazıyor, frontend'e WebSocket endpoint sunmuyor. Frontend doğrudan Binance WS'e mi bağlanmalı (mevcut plandaki yaklaşım) yoksa backend'e bir WS gateway eklenmeli mi?

> [!WARNING]
> **Stripe Frontend Entegrasyonu:** Payment service Stripe kullanıyor. Frontend'de Stripe Elements (kart formu) entegrasyonu yapılacak mı? Bu durumda `NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY` gerekecektir.

---

## 9. Verification Plan

### Automated Tests
- `npm run build` — TypeScript derleme hatası kontrolü
- `npm run lint` — ESLint kuralları
- Browser ile tüm sayfaların responsive kontrolü (320px, 768px, 1280px)

### Manual Verification
- Tüm API endpoint'lerine bağlantı testi (backend çalışırken)
- Login/Register flow end-to-end testi
- Wallet deposit/withdraw flow testi
- Order placement flow testi
- WebSocket bağlantı ve canlı fiyat akışı testi
- Dark/Light mode geçiş testi
- Mobile responsive kontrol (hamburger menu, layout)
