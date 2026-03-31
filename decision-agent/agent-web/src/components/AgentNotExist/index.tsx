import GradientContainer from '@/components/GradientContainer';
import EmptyIcon from '@/assets/images/empty.svg';
import { Button } from 'antd';
import { useNavigate } from 'react-router-dom';
import intl from 'react-intl-universal';

const AgentNotExist = () => {
  const navigate = useNavigate();
  return (
    <GradientContainer className="dip-full dip-flex-center">
      <div className="dip-flex-column-center">
        <EmptyIcon />
        <div>{intl.get('dataAgent.agentNotExist.title')}</div>
        <div className="dip-mt-12 dip-text-color-45">{intl.get('dataAgent.agentNotExist.description')}</div>
        <Button onClick={() => navigate(-1)} className="dip-mt-16">
          {intl.get('dataAgent.agentNotExist.back')}
        </Button>
      </div>
    </GradientContainer>
  );
};

export default AgentNotExist;
