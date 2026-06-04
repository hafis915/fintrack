import axios from 'axios'
import { http } from './client'
import type { Transaction } from './transactions'

export interface ReceiptDraft {
  amount: number
  merchant: string
  category_id?: string
  category_hint: string
  confidence: number
}

export interface ConfirmReceiptInput {
  file: File
  amount: number
  merchant: string
  category_id: string
  note?: string
  transacted_at: string // RFC3339
  ai_confidence: number
}

interface Envelope<T> {
  data?: T
  error?: { code: string; message: string }
}

function unwrap<T>(p: Promise<{ data: Envelope<T> }>): Promise<T> {
  return p
    .then(({ data }) => {
      if (data.data === undefined) throw new Error(data.error?.message ?? 'request failed')
      return data.data
    })
    .catch((err) => {
      if (axios.isAxiosError(err) && err.response?.data?.error?.message) {
        throw new Error(err.response.data.error.message)
      }
      throw err
    })
}

export function analyzeReceipt(file: File): Promise<ReceiptDraft> {
  const form = new FormData()
  form.append('image', file)
  // Let axios/browser set the multipart boundary; do not set Content-Type.
  return unwrap(http.post<Envelope<ReceiptDraft>>('/v1/receipts/analyze', form))
}

export function confirmReceipt(input: ConfirmReceiptInput): Promise<Transaction> {
  const form = new FormData()
  form.append('image', input.file)
  form.append('amount', String(input.amount))
  form.append('merchant', input.merchant)
  form.append('category_id', input.category_id)
  if (input.note) form.append('note', input.note)
  form.append('transacted_at', input.transacted_at)
  form.append('ai_confidence', String(input.ai_confidence))
  return unwrap(http.post<Envelope<Transaction>>('/v1/receipts/confirm', form))
}
