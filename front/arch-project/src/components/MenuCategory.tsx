import React from 'react';
import { useNavigate } from 'react-router-dom';
import '../styles/MenuCategory.css';

export interface MenuItem {
  id: number;
  name: string;
  price: number;
  description?: string;
  image?: string;
}

export interface MenuCategoryProps {
  id: number;
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
