# 🚀 TradeOff: Yeni Nesil Kripto Alım-Satım Platformu

TradeOff, modern mikroservis mimarisi ve yüksek performanslı teknolojilerle geliştirilmiş, uçtan uca bir kripto para ticaret platformudur. Hem kurumsal hem de bireysel kullanıcılar için düşük gecikmeli, güvenilir ve estetik bir deneyim sunar.

---

## 📺 Proje Tanıtım Videosu

Projenin çalışma mantığını ve arayüz detaylarını aşağıdaki videodan izleyebilirsiniz:

<div align="center">
  <video src="frontend/public/Ekran%20Kaydı%202026-05-03%20161105.mp4" width="100%" controls></video>
</div>

---

## ✨ Temalar (Koyu & Açık Mod)

TradeOff, kullanıcı tercihine göre dinamik olarak değiştirilebilen profesyonel bir karanlık ve aydınlık tema desteği sunar.

| 🌙 Koyu Tema | ☀️ Açık Tema |
| :---: | :---: |
| ![Koyu Tema](frontend/public/Ekran%20görüntüsü%202026-05-03%20161921.png) | ![Açık Tema](frontend/public/Ekran%20görüntüsü%202026-05-03%20161942.png) |

---

## 🛠️ Teknoloji Yığını

### 🏗️ Backend (Go Mikroservisleri)
- **Dil:** Go 1.22+
- **Mimari:** Mikroservisler, Clean Architecture, Event-Driven Design
- **Mesaj Kuyruğu:** Apache Kafka (Hizmetler arası asenkron iletişim)
- **Veritabanı:** PostgreSQL (İşlemsel veriler), Redis (Hızlı önbellekleme)
- **İletişim:** gRPC (Yüksek hızlı servisler arası çağrılar), REST & WebSockets (Client iletişimi)
- **Konteynerleştirme:** Docker & Docker Compose

### 🎨 Frontend (Next.js & React)
- **Framework:** Next.js 14+ (App Router)
- **Stil:** Tailwind CSS (Modern ve duyarlı tasarım)
- **Durum Yönetimi:** Zustand
- **Grafikler:** Lightweight Charts (Gerçek zamanlı TradingView grafikleri)
- **Animasyonlar:** Framer Motion
- **Veri Akışı:** Axios (REST), Native WebSockets (Canlı fiyat takibi)

---

## 🧩 Sistem Mimarisi

Sistem, Kafka tabanlı bir olay akışı ile birbirine bağlanan bağımsız mikroservislerden oluşur:

- **Broker Backend:** Ana API ağ geçidi, cüzdan yönetimi ve emir yönetimi.
- **Matching Engine:** Yüksek hızlı emir eşleştirme motoru.
- **Payment Service:** Yatırma ve çekme işlemleri için gRPC tabanlı servis.
- **Ingestion Service:** Dış borsalardan gerçek zamanlı piyasa verisi toplayıcı.

---

## 🚀 Kurulum ve Çalıştırma

### 1. Backend Kurulumu
Backend tarafındaki altyapı servislerini (PostgreSQL, Kafka, Redis) ayağa kaldırmak için:

```bash
cd backend/broker-backend
docker-compose up -d
```

Her bir servisi ilgili klasörüne giderek çalıştırabilirsiniz:
```bash
go run cmd/server/main.go
```

### 2. Frontend Kurulumu
Gerekli paketleri yüklemek ve geliştirme sunucusunu başlatmak için:

```bash
cd frontend
npm install
npm run dev
```

---

## 🎯 Temel Özellikler

- **Gerçek Zamanlı İşlemler:** WebSockets üzerinden anlık fiyat güncellemeleri ve emir takibi.
- **Gelişmiş Grafik Paneli:** TradingView altyapısıyla profesyonel teknik analiz imkanı.
- **Güvenli Cüzdan:** `DECIMAL(18,8)` hassasiyetinde bakiye yönetimi ve işlemsel doğruluk.
- **Hızlı Emir Eşleştirme:** Kafka destekli düşük gecikmeli emir işleme süreci.
- **Responsive Tasarım:** Mobil, tablet ve masaüstü cihazlarla tam uyumlu arayüz.

---

## 🛡️ Güvenlik ve Kararlılık

- **Outbox Pattern:** Veritabanı ve Kafka mesajları arasındaki veri tutarlılığını garanti eder.
- **Atomik İşlemler:** Tüm finansal hareketler PostgreSQL izolasyon seviyeleriyle korunur.
- **JWT Yetkilendirme:** Güvenli oturum yönetimi ve API erişimi.

---

*TradeOff - Finansal teknolojinin geleceği için tasarlandı.*
