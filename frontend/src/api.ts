import type {
  Alert,
  Position,
  Commodity,
  PricePoint,
  AnalysisResponse,
  MarketDataRow,
} from "./types";

const BASE_URL = import.meta.env.VITE_API_URL || "http://localhost:8000";

let currentUserId = "1";

export function setCurrentUserId(id: string) {
  currentUserId = id;
}

export function getCurrentUserId(): string {
  return currentUserId;
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      "X-User-ID": currentUserId,
      ...options?.headers,
    },
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `Request failed: ${res.status}`);
  }
  return res.json();
}

export async function listAlerts(
  status?: string,
  commodityCode?: string
): Promise<Alert[]> {
  const params = new URLSearchParams();
  if (status) params.set("status", status);
  if (commodityCode) params.set("commodity_code", commodityCode);
  const qs = params.toString();
  return request<Alert[]>(`/alerts${qs ? `?${qs}` : ""}`);
}

export async function createAlert(data: {
  commodity_code: string;
  condition: string;
  threshold_price: number;
  notes: string;
}): Promise<Alert> {
  return request<Alert>("/alerts", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function listCommodities(): Promise<Commodity[]> {
  return request<Commodity[]>("/commodities");
}

export async function listPositions(): Promise<Position[]> {
  return request<Position[]>("/positions");
}

export async function getCurrentPrices(): Promise<PricePoint[]> {
  return request<PricePoint[]>("/prices");
}

export async function getMarketData(
  location?: string,
  startDate?: string,
  endDate?: string
): Promise<MarketDataRow[]> {
  const params = new URLSearchParams();
  if (location) params.set("location", location);
  if (startDate) params.set("start_date", startDate);
  if (endDate) params.set("end_date", endDate);
  const qs = params.toString();
  return request<MarketDataRow[]>(`/market-data${qs ? `?${qs}` : ""}`);
}

export async function analyzePositions(): Promise<AnalysisResponse> {
  return request<AnalysisResponse>("/analyze-positions-market", {
    method: "POST",
  });
}
