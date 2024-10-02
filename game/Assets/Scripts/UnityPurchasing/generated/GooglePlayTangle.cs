// WARNING: Do not modify! Generated file.

namespace UnityEngine.Purchasing.Security {
    public class GooglePlayTangle
    {
        private static byte[] data = System.Convert.FromBase64String("r42UnIJem5tL1Mr8jmN5YKoLpk1PE2hVJo6o7QGIH96Pg8jc1hrL/kOGCVD9GVPcT5ggICx8k8uPsCCwo4IeUkjXtSLwFlndygiN6j1UfNuxL300OX9WBbEU9GrtyGwbirOm8t4aTv66P/0bL6ILhLENSa4anFdOVLoDhdwTA2iXVb8qc0cgRkKwvfpPz+nfG0xW1NxyCxwr3NiElC9YpoIX8/dC0FVaoFo9G2W+hsEleTKHLetqZs8dY+/D5iOOamKJ+g3/ZmRC8HNQQn90e1j0OvSFf3Nzc3dycQTrJGBIiZdRdATjU7i/WZv4u0Jp8HN9ckLwc3hw8HNzcu8xuYwOuI1aSwkb1ZiNvsdcVXCxKT4iXGs5dDR/newnW30e13Bxc3Jz");
        private static int[] order = new int[] { 4,13,2,7,8,7,9,11,12,9,12,11,13,13,14 };
        private static int key = 114;

        public static readonly bool IsPopulated = true;

        public static byte[] Data() {
        	if (IsPopulated == false)
        		return null;
            return Obfuscator.DeObfuscate(data, order, key);
        }
    }
}
