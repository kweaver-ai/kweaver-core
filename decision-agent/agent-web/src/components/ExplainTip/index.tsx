/**
 * tip解释
 */
import React from 'react';
import BaseTip, { BaseTipProps } from './BaseTip';

type ExplainTipInterface = React.FC<BaseTipProps>;

const ExplainTip = BaseTip as ExplainTipInterface;

export default ExplainTip;
