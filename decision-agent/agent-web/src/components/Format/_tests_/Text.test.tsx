import React from 'react';
import { shallow } from 'enzyme';

import Text from '../Text';

const init = (props: any) => shallow(<Text {...props} />);

describe('Text', () => {
  it('class name ad-format-text is exists', () => {
    const wrapper = init({});

    expect(wrapper.find('.ad-format-text').exists()).toBe(true);
  });
});
