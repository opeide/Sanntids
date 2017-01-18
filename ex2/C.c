// gcc -std=gnu99 -Wall -g -o oppgave4_c oppgave4.c -lpthread

#include <pthread.h>
#include <stdio.h>

int i = 0;

void* thread1_func(){
    for(int j = 0; j < 1000000; ++j) i++;
    return NULL;
}

void* thread2_func(){
    for(int j = 0; j < 1000000; ++j) i--;
    return NULL;
}

int main(){
    pthread_t thread1;
    pthread_create(&thread1, NULL, thread1_func, NULL);

    pthread_t thread2;
    pthread_create(&thread2, NULL, thread2_func, NULL);

    pthread_join(thread1, NULL);
    pthread_join(thread2, NULL);
    printf("num: %d\n", i);

    return 0;
}

/*// Note the return type: void*
void* someThreadFunction(){
    printf("Hello from a thread!\n");
    return NULL;
}



int main(){
    pthread_t someThread;
    pthread_create(&someThread, NULL, someThreadFunction, NULL);
    // Arguments to a thread would be passed here ---------^
    
    pthread_join(someThread, NULL);
    printf("Hello from main!\n");
    return 0;
    
}*/
