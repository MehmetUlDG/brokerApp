export function StatsSection() {
  const stats = [
    { label: 'Günlük Hacim', value: '$2.4B+' },
    { label: 'Aktif Kullanıcı', value: '1.2M+' },
    { label: 'Desteklenen Kripto', value: '150+' },
    { label: 'Ortalama İşlem Süresi', value: '<50ms' },
  ];

  return (
    <section className="border-y border-[var(--border)] bg-[var(--bg-secondary)] py-16" id="stats">
      <div className="container mx-auto px-4">
        <div className="grid grid-cols-2 gap-8 md:grid-cols-4">
          {stats.map((stat, i) => (
            <div key={i} className="text-center">
              <div className="text-4xl font-extrabold text-[var(--text-primary)] md:text-5xl">{stat.value}</div>
              <div className="mt-2 text-sm font-medium text-[var(--text-secondary)]">{stat.label}</div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
