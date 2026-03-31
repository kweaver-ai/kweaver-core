export interface Entity {
  key: string;
  title: string;
  alias: string;
  name: string;
  color?: string;
  properties?: any[];
  properties_index?: string[];
}

export interface KnowledgeNetwork {
  id: string;
  name: string;
  spaceId?: string;
}

export interface SelectedData {
  spaceId?: string;
  networkId: string;
  entities: string[];
  properties?: Record<string, string[]>;
} 