import React from 'react';

type ContainerIsHideType = {
  isVisible?: boolean;
  placeholder?: any;
  children: React.ReactNode;
};

const ContainerIsVisible = (props: Omit<ContainerIsHideType, 'isVisible' | 'placeholder'>) => {
  return <React.Fragment>{props.children}</React.Fragment>;
};

export default (props: ContainerIsHideType) => {
  const { isVisible, placeholder, ...other } = props;
  if (!isVisible) return placeholder || null;
  return <ContainerIsVisible {...other} />;
};
