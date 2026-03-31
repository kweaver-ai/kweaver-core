import React from 'react';
import { shallow } from 'enzyme';

import Button from '../Button';

const init = (props: any) => shallow(<Button {...props} />);

describe('Button', () => {
  it('class name ad-format-button is exists', () => {
    const wrapper = init({});
    expect(wrapper.exists()).toBe(true);
  });
});
