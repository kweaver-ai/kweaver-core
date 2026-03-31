import React from 'react';
import BasicInfo from './BasicInfo';
import RoleInstruction from './RoleInstruction';

const Sidebar: React.FC = () => {
  return (
    <div className="dip-w-100 dip-h-100 dip-position-r dip-overflowY-auto dip-bg-white">
      <BasicInfo />
      <RoleInstruction />
    </div>
  );
};

export default Sidebar;
