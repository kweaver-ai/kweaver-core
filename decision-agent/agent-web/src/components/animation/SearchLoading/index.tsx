import React from 'react';
import Lottie from 'lottie-react';
import searchLoadingJSON from './searchLoading.json';

const SearchLoading = () => {
  return <Lottie style={{ width: 22, height: 22 }} animationData={searchLoadingJSON} loop />;
};

export default SearchLoading;
