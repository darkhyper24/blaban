import React from 'react';
import { useNavigate } from 'react-router-dom';
import '../styles/MenuCategory.css';

// Adjusted to match the API response structure
export interface MenuItem {
  id: string | number;
  name: string;
  price: number;
  effective_price?: number;
  is_available?: boolean;
  quantity?: number;
  has_discount?: boolean;
  discount_value?: number;
  description?: string;
  image?: string;
}

export interface MenuCategoryProps {
  id: string | number;
  name: string;
  items: MenuItem[];
  icon?: string;
  letter?: string;
}

const MenuCategory: React.FC<MenuCategoryProps> = ({ id, name, items, icon, letter }) => {
  const navigate = useNavigate();

  const goToCategoryPage = () => {
    navigate(`/category/${id}`, { state: { categoryName: name, items } });
  };

  return (
    <div className="category-card" onClick={goToCategoryPage}>
      <div className="category-icon">
        {icon ? (
          <img src={icon} alt={name} className="category-image" />
        ) : (
          <div className="category-letter">{letter || name.charAt(0)}</div>
        )}
      </div>
      <h3>{name}</h3>
    </div>
  );
};

export default MenuCategory;
