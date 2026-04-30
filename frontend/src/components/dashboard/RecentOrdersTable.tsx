'use client';

import { Card } from '@/components/ui/Card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { useTradeStore } from '@/stores/tradeStore';
import { format } from 'date-fns';

export function RecentOrdersTable() {
  const orders = useTradeStore((state) => state.orders);
  
  // Sadece ilk 5 emri göster
  const recentOrders = orders.slice(0, 5);

  return (
    <Card className="overflow-hidden">
      <div className="border-b border-[var(--border)] p-4">
        <h3 className="font-bold text-[var(--text-primary)]">Son Emirler</h3>
      </div>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Tarih</TableHead>
            <TableHead>Sembol</TableHead>
            <TableHead>Tür / Yön</TableHead>
            <TableHead>Fiyat</TableHead>
            <TableHead>Miktar</TableHead>
            <TableHead>Durum</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {recentOrders.length === 0 ? (
            <TableRow>
              <TableCell colSpan={6} className="text-center text-[var(--text-muted)] py-8">
                Henüz emir bulunmuyor.
              </TableCell>
            </TableRow>
          ) : (
            recentOrders.map((order) => (
              <TableRow key={order.id}>
                <TableCell className="text-[var(--text-secondary)]">
                  {format(new Date(order.created_at), 'dd MMM yyyy HH:mm')}
                </TableCell>
                <TableCell className="font-medium text-[var(--text-primary)]">{order.symbol}</TableCell>
                <TableCell>
                  <div className="flex items-center gap-2">
                    <span className="text-xs font-medium text-[var(--text-secondary)]">{order.type}</span>
                    <Badge variant={order.side === 'BUY' ? 'success' : 'danger'}>
                      {order.side}
                    </Badge>
                  </div>
                </TableCell>
                <TableCell className="font-mono text-[var(--text-primary)]">${parseFloat(order.price).toLocaleString()}</TableCell>
                <TableCell className="font-mono text-[var(--text-primary)]">{order.quantity}</TableCell>
                <TableCell>
                  <Badge
                    variant={
                      order.status === 'COMPLETED' ? 'success' :
                      order.status === 'PENDING' ? 'warning' :
                      order.status === 'FAILED' ? 'danger' : 'neutral'
                    }
                  >
                    {order.status}
                  </Badge>
                </TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </Card>
  );
}
