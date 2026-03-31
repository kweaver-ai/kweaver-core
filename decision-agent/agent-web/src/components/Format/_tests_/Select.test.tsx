import React from 'react';
import { shallow } from 'enzyme';

import Select from '../Select';

const init = (props: any) => shallow(<Select {...props} />);

describe('Select', () => {
  it('class name ad-format-select is exists', () => {
    const wrapper = init({});

    expect(wrapper.find('.ad-format-select').exists()).toBe(true);
  });
});
