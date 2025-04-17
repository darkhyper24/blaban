import React from 'react';
import { MenuCategoryProps } from '../components/MenuCategory';
import MenuCategory from '../components/MenuCategory';
import '../styles/Home.css';

// Import images from assets
import qashtouta from '../assets/2a4tota.PNG';
import qashtooza from '../assets/2a4toza.PNG';
import omAli from '../assets/om-3li.PNG';
import koshary from '../assets/koshary.PNG';

const Home: React.FC = () => {
  // Sample menu data based on the image provided
  const menuCategories: MenuCategoryProps[] = [
    {
      id: 1,
      name: 'قشطوطة',
      icon: qashtouta,
      items: [
        { id: 1, name: 'قشطوطة كراميل', price: 55 },
        { id: 2, name: 'قشطوطة مانجا', price: 70 },
        { id: 3, name: 'قشطوطة مكسرات', price: 80 },
        { id: 4, name: 'قشطوطة لوتس', price: 80 },
        { id: 5, name: 'قشطوطة قشطة', price: 55 },
        { id: 6, name: 'قشطوطة رز ب لبن مكسرات', price: 80 },
        { id: 7, name: 'قشطوطة رز ب لبن لوتس', price: 85 },
        { id: 8, name: 'قشطوطة رز ب لبن مانجا', price: 70 },
        { id: 9, name: 'قشطوطة رز ب لبن كريمة', price: 55 },
        { id: 10, name: 'قشطوطة سوبر لوكس', price: 100 },
      ]
    },
    {
      id: 2,
      name: 'حلويات شرقية',
      icon: omAli,
      items: [
        { id: 11, name: 'بسبوسة', price: 45 },
        { id: 12, name: 'كنافة', price: 50 },
        { id: 13, name: 'بقلاوة', price: 60 }
      ]
    },
    {
      id: 3,
      name: 'مشروبات',
      icon: qashtooza,
      items: [
        { id: 14, name: 'عصير مانجو', price: 30 },
        { id: 15, name: 'عصير فراولة', price: 25 },
        { id: 16, name: 'قهوة', price: 20 }
      ]
    },
    {
      id: 4,
      name: 'وجبات',
      icon: koshary,
      items: [
        { id: 17, name: 'فطور لبناني', price: 120 },
        { id: 18, name: 'سندويشات', price: 65 },
        { id: 19, name: 'مقبلات', price: 40 }
      ]
    }
  ];

  return (
    <div className="home-container">
      <div className="welcome-section">
        <h1>مرحباً بكم في بلبن</h1>
        <p>من دنيا حلوة</p>
      </div>
      
      <div className="categories-container">
        {menuCategories.map(category => (
          <MenuCategory 
            key={category.id}
            id={category.id}
            name={category.name}
            items={category.items}
            icon={category.icon}
          />
        ))}
      </div>
    </div>
  );
};

export default Home;