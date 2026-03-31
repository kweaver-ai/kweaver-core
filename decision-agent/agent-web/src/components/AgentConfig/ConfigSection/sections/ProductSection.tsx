import React, { useState, useEffect } from 'react';
import { Select, message } from 'antd';
import intl from 'react-intl-universal';
import { useAgentConfig } from '../../AgentConfigContext';
import { getProductList } from '@/apis/agent-factory';
import ProductIcon from '@/assets/icons/product.svg';
import SectionPanel from '../../common/SectionPanel';

const { Option } = Select;

interface SelectOption {
  value: string;
  label: string;
}

const ProductSection: React.FC = () => {
  const { state, actions } = useAgentConfig();
  const [selectedProduct, setSelectedProduct] = useState<string | number>(state.product_key || '');
  const [productOptions, setProductOptions] = useState<SelectOption[]>([]);
  const [isLoadingProducts, setIsLoadingProducts] = useState(false);
  const [isExpanded, setIsExpanded] = useState(false);

  // 检查是否可编辑产品配置
  const canEditProduct = actions.canEditField('product');

  // 获取产品列表
  const fetchProducts = async () => {
    setIsLoadingProducts(true);
    try {
      const response = await getProductList();
      if (response && response.entries && response.entries.length > 0) {
        const productOptions = response.entries.map((product: any) => ({
          value: product.key.toString(),
          label: product.name,
        }));
        setProductOptions(productOptions);

        if (productOptions.length > 0 && !selectedProduct) {
          // 如果有可用产品但尚未选择任何产品，则设置默认产品
          handleProductChange(productOptions[0].value);
        } else if (selectedProduct && !response.entries.find((product: any) => product.key === selectedProduct)) {
          // 如果已选的产品不存在了，则设置为空
          setSelectedProduct('');
          actions.updateProductId('');
        }
      }
    } catch (error) {
      console.error('Failed to fetch products:', error);
      message.error(intl.get('dataAgent.fetchProductListFail'));
    } finally {
      setIsLoadingProducts(false);
    }
  };

  // 处理产品变更
  const handleProductChange = (value: string) => {
    if (!canEditProduct) return;
    setSelectedProduct(value);
    actions.updateProductId(value);
  };

  // 组件挂载时获取产品列表
  useEffect(() => {
    fetchProducts();
  }, []);

  return (
    <SectionPanel
      title={intl.get('dataAgent.config.associatedProduct')}
      description={intl.get('dataAgent.config.pleaseSelectAssociatedProduct')}
      isExpanded={isExpanded}
      onToggle={() => setIsExpanded(!isExpanded)}
      icon={<ProductIcon />}
      className="dip-border-line-b"
    >
      <div className="product-config">
        <Select
          style={{ width: '100%' }}
          value={selectedProduct.toString() || '---'}
          onChange={handleProductChange}
          loading={isLoadingProducts}
          disabled={!canEditProduct}
          placeholder={intl.get('dataAgent.config.pleaseSelectProduct')}
          defaultActiveFirstOption={false}
          status={!selectedProduct ? 'error' : ''}
        >
          {productOptions.map(option => (
            <Option key={option.value} value={option.value}>
              {option.label}
            </Option>
          ))}
        </Select>
      </div>
    </SectionPanel>
  );
};

export default ProductSection;
