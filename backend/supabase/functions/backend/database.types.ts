export type Json =
  | string
  | number
  | boolean
  | null
  | { [key: string]: Json | undefined }
  | Json[];

export type Database = {
  public: {
    Tables: {
      device_accounts: {
        Row: {
          id: string;
          device_identifier_hash: string;
          access_code_hash: string;
          access_code_counter: number;
          login_email: string;
          created_at: string;
          last_login_at: string | null;
        };
        Insert: {
          id: string;
          device_identifier_hash: string;
          access_code_hash: string;
          access_code_counter: number;
          login_email: string;
          created_at?: string;
          last_login_at?: string | null;
        };
        Update: {
          id?: string;
          device_identifier_hash?: string;
          access_code_hash?: string;
          access_code_counter?: number;
          login_email?: string;
          created_at?: string;
          last_login_at?: string | null;
        };
        Relationships: [];
      };
      scores: {
        Row: {
          user_id: string;
          score: number;
          max_combo: number;
          created_at: string;
          updated_at: string;
        };
        Insert: {
          user_id: string;
          score: number;
          max_combo?: number;
          created_at?: string;
          updated_at?: string;
        };
        Update: {
          user_id?: string;
          score?: number;
          max_combo?: number;
          created_at?: string;
          updated_at?: string;
        };
        Relationships: [
          {
            foreignKeyName: "scores_user_id_fkey";
            columns: ["user_id"];
            isOneToOne: true;
            referencedRelation: "device_accounts";
            referencedColumns: ["id"];
          },
        ];
      };
    };
    Views: {
      leaderboard: {
        Row: {
          rank: number;
          user_id: string;
          score: number;
          max_combo: number;
          updated_at: string;
        };
        Relationships: [];
      };
    };
    Functions: Record<string, never>;
    Enums: Record<string, never>;
    CompositeTypes: Record<string, never>;
  };
};

export type ScoreRow = Database["public"]["Tables"]["scores"]["Row"];
