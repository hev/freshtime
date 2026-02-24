import type { HttpClient } from "./http.ts";

export interface InvoiceLine {
  type: 0;
  name: string;
  description: string;
  qty: string;
  unit_cost: { amount: string; code: string };
}

export interface CreateInvoiceRequest {
  invoice: {
    customerid: number;
    create_date: string;
    lines: InvoiceLine[];
    status: 1 | 2;
    notes?: string;
    [key: string]: unknown;
  };
}

export interface InvoiceResponse {
  invoiceid: number;
  invoice_number: string;
  amount: { amount: string; code: string };
  links: { client_view: string } | null;
  v3_status: string;
}

export async function createInvoice(
  http: HttpClient,
  accountId: string,
  request: CreateInvoiceRequest
): Promise<InvoiceResponse> {
  const data = await http.post<{ response: { result: { invoice: InvoiceResponse } } }>(
    `/accounting/account/${accountId}/invoices/invoices`,
    request
  );
  return data.response.result.invoice;
}

export async function getShareLink(
  http: HttpClient,
  accountId: string,
  invoiceId: number
): Promise<string | null> {
  try {
    const data = await http.get<{ response: { result: { share_link: string } } }>(
      `/accounting/account/${accountId}/invoices/invoices/${invoiceId}/share_link`
    );
    return data.response.result.share_link;
  } catch {
    return null;
  }
}
