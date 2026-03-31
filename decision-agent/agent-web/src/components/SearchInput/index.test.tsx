import React from 'react';
import { shallow } from 'enzyme';

import SearchInput from './index';

const init = (props = {}) => shallow(<SearchInput {...props} />);

describe('SearchInput', () => {
  it('init', () => {
    const wrapper = init();
    expect(wrapper.exists()).toBe(true);
  });
});
