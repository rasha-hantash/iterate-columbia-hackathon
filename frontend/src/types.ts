export interface Alert {
  id: number;
  client_id: number;
  user_id: number;
  commodity_id: number;
  commodity_code: string;
  commodity_name: string;
  condition: string;
  threshold_price: number;
  status: string;
  notes: string;
  triggered_count: number;
  last_triggered_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface Position {
  id: number;
  client_id: number;
  user_id: number;
  commodity_id: number;
  commodity_code: string;
  commodity_name: string;
  volume: number;
  direction: string;
  entry_price: number;
}

export interface Commodity {
  id: number;
  code: string;
  name: string;
  unit: string;
}

export interface PricePoint {
  commodity_id: number;
  commodity_code: string;
  commodity_name: string;
  price: number;
  recorded_at: string;
}

export interface AlertSuggestion {
  commodity_code: string;
  condition: string;
  threshold_price: number;
  notes: string;
}

export interface AnalysisResponse {
  reasoning: string;
  suggestions: AlertSuggestion[];
}

export interface MarketDataRow {
  id: number;
  report_date: string;
  location: string;
  commodity: string;
  variety: string;
  origin: string;
  item_size: string;
  low_price: number | null;
  high_price: number | null;
  properties: string;
}
