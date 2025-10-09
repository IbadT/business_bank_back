import { Controller, Get } from '@nestjs/common';

@Controller()
export class AppController {
  constructor() {}

  @Get('/health_check')
  healthCheck(): Record<string, string> {
    return {
      status: 'Microservice maska is running',
    };
  }
}
