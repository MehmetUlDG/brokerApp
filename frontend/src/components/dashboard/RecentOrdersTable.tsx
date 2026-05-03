'use client';

import { Card } from '@/components/ui/Card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { useOrders } from '@/hooks/useOrders';
import { Skeleton } from '@/components/ui/Skeleton';
import { format } from 'date-fns';

export function RecentOrdersTable() {
  const { orders, ordersLoading } = useOrders();

  // Sadece ilk 5 emri göster
  const recentOrders = orders.slice(0, 5);

  return (
    <Card className="overflow-hidden">
      <div className="border-b border-[var(--border)] p-4 flex items-center justify-between">
        <h3 className="font-bold text-[var(--text-primary)]">Son Emirler</h3>
        {ordersLoading && (
          <span className="text-xs text-[var(--text-muted)] animate-pulse">Güncelleniyor...</span>
        )}
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
          {ordersLoading && orders.length === 0 ? (
            <>
              {[...Array(3)].map((_, i) => (
                <TableRow key={i}>
                  {[...Array(6)].map((_, j) => (
                    <TableCell key={j}><Skeleton className="h-4 w-full" /></TableCell>
                  ))}
                </TableRow>
              ))}
            </>
          ) : recentOrders.length === 0 ? (
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
