import React from 'react';
import { useParams, useLocation, Link } from 'react-router-dom';
import '../styles/CategoryItems.css';

interface MenuItem {
  id: number;
  name: string;
  price: number;
  description?: string;
  image?: string;
}

const CategoryItems: React.FC = () => {
  // We get the categoryId from params, but primarily use the state passed via navigation
  // The categoryId could be used for fetching data from an API if needed
  const { categoryId } = useParams<{ categoryId: string }>();
  const location = useLocation();
  const { categoryName, items } = location.state || { categoryName: '', items: [] };
  
  // Log the category ID for debugging purposes
  console.log(`Viewing category ID: ${categoryId}`);
  

  if (!items || items.length === 0) {
    return (
      <div className="category-items-container">
        <div className="category-header">
          <Link to="/" className="back-button">← Back to Categories</Link>
          <h1>No items found</h1>
        </div>
      </div>
    );
  }

  return (
    <div className="category-items-container">
      <div className="category-header">
        <Link to="/" className="back-button">← Back to Categories</Link>
        <h1>{categoryName}</h1>
      </div>
      
      <div className="items-grid">
        {items.map((item: MenuItem) => (
          <div key={item.id} className="item-card">
            {item.image && (
              <div className="item-image-container">
                <img src={item.image} alt={item.name} className="item-image" />
              </div>
            )}
            <div className="item-details">
              <h2>{item.name}</h2>
              <p className="item-price">{item.price} LE</p>
              {item.description && <p className="item-description">{item.description}</p>}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default CategoryItems;
