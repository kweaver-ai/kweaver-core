import dayjs from 'dayjs';

export const formatTime = (time: string | number) => (time ? dayjs(time).format('YYYY-MM-DD HH:mm:ss') : '- -');

export const formatTimeSlash = (time: string | number) => (time ? dayjs(time).format('YYYY/MM/DD HH:mm:ss') : '- -');
