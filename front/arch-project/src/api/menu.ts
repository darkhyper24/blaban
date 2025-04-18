import { apiConfig } from './config';

export interface MenuItem {
  id: string;
  name: string;
  price: number;
  effective_price: number;
  is_available: boolean;
  quantity: number;
  has_discount: boolean;
  discount_value: number;
}

export interface MenuCategory {
  id: string;
  name: string;
  picture?: string;
  items?: MenuItem[];
}

const menuApi = {
  // Get all categories
  getCategories: async (): Promise<MenuCategory[]> => {
    try {
      const response = await fetch(`${apiConfig.menuUrl}/api/categories`);
      
      if (!response.ok) {
        throw new Error('Failed to fetch categories');
      }
      
      const data = await response.json();
      return data.categories || [];
    } catch (error) {
      console.error('Error fetching categories:', error);
      return [];
    }
  },
  
  // Get menu with items grouped by category
  getMenu: async (): Promise<MenuCategory[]> => {
    try {
      const response = await fetch(`${apiConfig.menuUrl}/api/menu`);
      
      if (!response.ok) {
        throw new Error('Failed to fetch menu');
      }
      
      const data = await response.json();
      return data.menu || [];
    } catch (error) {
      console.error('Error fetching menu:', error);
      return [];
    }
  },
  
  // Get a specific menu item by ID
  getMenuItem: async (id: string): Promise<MenuItem | null> => {
    try {
      const response = await fetch(`${apiConfig.menuUrl}/api/menu/${id}`);
      
      if (!response.ok) {
        throw new Error('Failed to fetch menu item');
      }
      
      const data = await response.json();
      return data.item || null;
    } catch (error) {
      console.error(`Error fetching menu item ${id}:`, error);
      return null;
    }
  },
  
  // Search menu items
  searchItems: async (query: string): Promise<MenuItem[]> => {
    try {
      const response = await fetch(`${apiConfig.menuUrl}/api/menu/search?q=${encodeURIComponent(query)}`);
      
      if (!response.ok) {
        throw new Error('Failed to search menu items');
      }
      
      const data = await response.json();
      return data.items || [];
    } catch (error) {
      console.error('Error searching menu items:', error);
      return [];
    }
  }
};

export default menuApi;
