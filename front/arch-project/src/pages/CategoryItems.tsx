import React, { useEffect } from 'react';
import { useParams, useLocation, Link } from 'react-router-dom';
import '../styles/CategoryItems.css';
import { MenuItem } from '../api/menu';



const CategoryItems: React.FC = () => {
  // We get the categoryId from params, but primarily use the state passed via navigation
  const { categoryId } = useParams<{ categoryId: string }>();
  const location = useLocation();
  const { categoryName, items } = location.state || { categoryName: '', items: [] };
  
  useEffect(() => {
    console.log(`Viewing category: ${categoryName}, ID: ${categoryId}`);
    console.log('Items received:', items);
  }, [categoryId, categoryName, items]);

  if (!items || items.length === 0) {
    return (
      <div className="category-items-container">
        <div className="category-header">
          <Link to="/" className="back-button">← Back to Categories</Link>
          <h1>No items found for {categoryName}</h1>
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
            <div className="item-details">
              <h2>{item.name}</h2>
              <p className="item-price">
                {item.has_discount ? (
                  <>
                    <span className="original-price">{item.price} LE</span>
                    <span className="discounted-price">{item.effective_price} LE</span>
                  </>
                ) : (
                  `${item.price} LE`
                )}
              </p>
              {item.is_available !== undefined && (
                <span className={`availability ${item.is_available ? 'available' : 'unavailable'}`}>
                  {item.is_available ? 'Available' : 'Out of stock'}
                </span>
              )}
              {item.quantity !== undefined && item.quantity > 0 && (
                <span className="quantity">Qty: {item.quantity}</span>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default CategoryItems;
