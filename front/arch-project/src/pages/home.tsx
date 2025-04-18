import React, { useEffect, useState } from 'react';
import { MenuCategoryProps } from '../components/MenuCategory';
import MenuCategory from '../components/MenuCategory';
import menuApi, { MenuCategory as ApiMenuCategory } from '../api/menu';
import '../styles/home.css';

const Home: React.FC = () => {
  const [menuCategories, setMenuCategories] = useState<MenuCategoryProps[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        
        // First, get the categories with their pictures
        const categoriesData = await menuApi.getCategories();
        console.log('Categories from API:', categoriesData);
        
        // Then, get the full menu with all items
        const menuData = await menuApi.getMenu();
        console.log('Menu data from API:', menuData);
        
        // Create a map of category IDs to their items
        const categoryItemsMap: Record<string, any[]> = {};
        menuData.forEach(category => {
          categoryItemsMap[category.id] = category.items || [];
        });
        
        // Combine the data: use categories for images, and menu for items
        const transformedData: MenuCategoryProps[] = categoriesData.map((category: ApiMenuCategory) => ({
          id: category.id,
          name: category.name,
          icon: category.picture, // Use the picture URL from categories API
          items: categoryItemsMap[category.id] || [] // Use items from menu API
        }));
        
        console.log('Combined data:', transformedData);
        
        setMenuCategories(transformedData);
        setError(null);
      } catch (err) {
        console.error('Error fetching data:', err);
        setError('Failed to load menu data. Please try again later.');
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  return (
    <div className="home-container">
      <div className="welcome-section">
        <h1>مرحباً بكم في بلبن</h1>
        <p>من دنيا حلوة</p>
      </div>
      
      {loading ? (
        <div className="loading-message">Loading menu...</div>
      ) : error ? (
        <div className="error-message">{error}</div>
      ) : (
        <div className="categories-container">
          {menuCategories.length > 0 ? (
            menuCategories.map(category => (
              <MenuCategory 
                key={category.id}
                id={category.id}
                name={category.name}
                items={category.items}
                icon={category.icon}
                letter={category.name.charAt(0)}
              />
            ))
          ) : (
            <div className="no-data-message">No menu categories available.</div>
          )}
        </div>
      )}
    </div>
  );
};

export default Home;