public class TestComplexity {
    
    public void simpleMethod() {
        System.out.println("Hello");
    }
    
    public void complexMethod(int x, int y) {
        if (x > 0) {
            for (int i = 0; i < 10; i++) {
                if (i % 2 == 0) {
                    System.out.println("Even");
                } else {
                    System.out.println("Odd");
                }
            }
        }
        
        while (y > 0) {
            if (y % 2 == 0 && x > 5) {
                break;
            }
            y--;
        }
        
        switch (x) {
            case 1:
                System.out.println("One");
                break;
            case 2:
                System.out.println("Two");
                break;
            default:
                System.out.println("Other");
        }
    }
    
    public void anotherComplexMethod() {
        try {
            riskyOperation();
        } catch (Exception e) {
            handleException(e);
        }
        
        int result = (x > 0) ? x * 2 : x / 2;
    }
    
    private void riskyOperation() throws Exception {
        throw new Exception("Test");
    }
    
    private void handleException(Exception e) {
        e.printStackTrace();
    }
}
