export const mainAppProps = {
  businessDomainID: 'bd_public',
  changeCustomPathComponent: () => {},
  config: {
    systemInfo: {
      location: {
        protocol: 'http:',
        hostname: 'localhost',
        port: '1101',
      },
    },
    getTheme: {
      normal: '#126ee3',
    },
  },
  prefix: '',
  language: {
    getLanguage: 'zh-CN',
  },
  token: {
    getToken: {
      access_token: 'ory_at_oql9Htv8ldzo3xiB1reC-urfprWy1rDKjURuYhFo7o4.Tr8uw11bGI_S8Bi4lEyb5ip89D4od8jnW6y1hAEsMlE',
    },
    onTokenExpired: () => {},
    refreshOauth2Token: () => {},
  },
  history: {
    navigateToMicroWidget: () => {},
  },
  userid: '4b91118a-6f67-11f0-b0dc-36fa540cff80',
};
