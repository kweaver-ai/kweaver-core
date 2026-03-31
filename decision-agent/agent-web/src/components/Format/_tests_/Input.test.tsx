import React from 'react';
import { shallow } from 'enzyme';

import Input from '../Input';

const init = (props: any) => shallow(<Input {...props} />);

describe('Input', () => {
  it('class name ad-format-input is exists', () => {
    const wrapper = init({});

    expect(wrapper.find('.ad-format-input').exists()).toBe(true);
  });
});
