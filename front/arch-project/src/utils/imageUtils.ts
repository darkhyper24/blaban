// /**
//  * Fixes issues with image URLs from the database
//  * - Removes double slashes (except in protocol)
//  * - Adds CORS proxy if needed
//  * - Ensures proper encoding
//  */
// export const fixImageUrl = (url: string | undefined): string => {
//   if (!url) return '';
  
//   try {
//     // Fix double slashes in URL path (not protocol)
//     let fixedUrl = url.replace(/(https?:\/\/)|(\/\/)/g, (match, protocol) => {
//       return protocol || '/';
//     });
    
//     // Create proper URL object to ensure it's valid
//     const urlObj = new URL(fixedUrl);
    
//     // Return the fixed URL
//     return urlObj.toString();
//   } catch (e) {
//     console.error('Invalid URL format:', url, e);
//     return '';
//   }
// }; 