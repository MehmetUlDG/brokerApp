'use client';

import { useEffect, useState } from 'react';
import { Card } from '@/components/ui/Card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { walletApi } from '@/lib/api/wallet';
import { Transaction } from '@/types/domain';
import { format } from 'date-fns';
import { Skeleton } from '@/components/ui/Skeleton';

export function TransactionHistory() {
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function loadTransactions() {
      try {
        const data = await walletApi.getTransactions();
        setTransactions(data);
      } catch (error) {
        console.error('Failed to load transactions', error);
      } finally {
        setLoading(false);
      }
    }
    loadTransactions();
  }, []);

  return (
    <Card className="overflow-hidden">
      <div className="border-b border-[var(--border)] p-4">
        <h3 className="font-bold text-[var(--text-primary)]">İşlem Geçmişi</h3>
      </div>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Tarih</TableHead>
            <TableHead>İşlem</TableHead>
            <TableHead>Tutar</TableHead>
            <TableHead>Durum</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {loading ? (
            Array.from({ length: 3 }).map((_, i) => (
              <TableRow key={i}>
                <TableCell><Skeleton className="h-4 w-24" /></TableCell>
                <TableCell><Skeleton className="h-4 w-16" /></TableCell>
                <TableCell><Skeleton className="h-4 w-20" /></TableCell>
                <TableCell><Skeleton className="h-4 w-16" /></TableCell>
              </TableRow>
            ))
          ) : transactions.length === 0 ? (
            <TableRow>
              <TableCell colSpan={4} className="text-center text-[var(--text-muted)] py-12">
                Henüz bir işleminiz bulunmuyor.
              </TableCell>
            </TableRow>
          ) : (
            transactions.map((tx) => (
              <TableRow key={tx.id}>
                <TableCell className="text-[var(--text-secondary)]">
                  {format(new Date(tx.created_at), 'dd MMM yyyy HH:mm')}
                </TableCell>
                <TableCell>
                  <div className="flex items-center gap-2">
                    <span className="font-medium text-[var(--text-primary)]">
                      {tx.type === 'DEPOSIT' ? 'Para Yatırma' : 'Para Çekme'}
                    </span>
                    <Badge variant={tx.type === 'DEPOSIT' ? 'success' : 'neutral'}>
                      {tx.type}
                    </Badge>
                  </div>
                </TableCell>
                <TableCell className="font-mono text-[var(--text-primary)]">
                  {tx.type === 'DEPOSIT' ? '+' : '-'}${parseFloat(tx.amount).toLocaleString()}
                </TableCell>
                <TableCell>
                  <Badge
                    variant={
                      tx.status === 'COMPLETED' ? 'success' :
                      tx.status === 'PENDING' ? 'warning' :
                      tx.status === 'FAILED' ? 'danger' : 'neutral'
                    }
                  >
                    {tx.status}
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
