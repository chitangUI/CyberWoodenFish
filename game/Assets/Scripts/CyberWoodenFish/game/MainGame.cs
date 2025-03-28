using System.Collections;
using System.Collections.Generic;
using System.Linq;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

namespace CyberWoodenFish.game
{
    public class MainGame : MonoBehaviour
    {

        private int _combo;

        private int _gongde;
        
        [SerializeField] private TMP_Text comboText;
        
        [SerializeField] private TMP_Text gongdeText;

        [SerializeField] private GameObject subText;

        [SerializeField] private AudioSource sb;

        [SerializeField] private Rigidbody chitang;
        
        private const float MoveSpeed = 200f; // Speed of the text movement
        
        private const float Lifetime = 0.5f; // How long the text stays visible
        
        private const float ComboHideDelay = 1f;

        private bool _boost;
        
        private float _lastComboUpdateTime;
        
        // Start is called once before the first execution of Update after the MonoBehaviour is created
        private void Start()
        { 
            _gongde = PlayerPrefs.GetInt("gongde"); // default 0
            _combo = 0;
            _lastComboUpdateTime = Time.time;

            _boost = PlayerPrefs.GetInt("boost") != 0;
        }

        // Update is called once per frame
        private void Update()
        {
            
            if (Input.GetMouseButtonDown(0)) { // 鼠标左键或触摸
                ApplyForceAtClick();
                if (_boost)
                {
                    _gongde -= 3;
                    _combo += 3;
                }
                _gongde -= 1;
                _combo += 1;
                _lastComboUpdateTime = Time.time;
                var text = Instantiate(subText, subText.transform.position, Quaternion.identity, subText.transform.parent);
                text.SetActive(true);
                StartCoroutine(MoveAndDestroyText(text));
            }
            
            // Update gongde text
            if (_gongde.ToString() != GetNumber(gongdeText))
            {
                PlayerPrefs.SetInt("gongde", _gongde);
                SetNumber(gongdeText, "当前功德", _gongde);
            }

            // Hide combo text if no update in the last second
            if (Time.time - _lastComboUpdateTime >= ComboHideDelay)
            {
                comboText.gameObject.SetActive(false);
            }
            else
            {
                // Show combo text if it's not visible and _combo is not zero
                if (_combo <= 0) return;
                comboText.gameObject.SetActive(true);
                if (_combo.ToString() != GetNumber(comboText))
                {
                    SetNumber(comboText, "Combo", _combo);
                }
            }
        }

        void ApplyForceAtClick() {
            // 创建从摄像机到鼠标位置的射线
            Ray ray = Camera.main.ScreenPointToRay(Input.mousePosition);
            RaycastHit hit;
            
            if (Physics.Raycast(ray, out hit)) {
                Rigidbody rb = hit.rigidbody;

                if (rb == null) return;
                Vector3 forcePoint = Vector3.Lerp(hit.point, rb.worldCenterOfMass, 0.3f);
                rb.AddForceAtPosition(ray.direction.normalized * 30f, forcePoint, ForceMode.Impulse);
            }
        }

        private IEnumerator MoveAndDestroyText(GameObject text)
        {

            sb.Play();
            
            var rectTransform = text.gameObject.GetComponent<RectTransform>();
            var elapsedTime = 0f;

            while (elapsedTime < Lifetime)
            {
                var yOffset = MoveSpeed * Time.deltaTime;
                rectTransform.anchoredPosition += new Vector2(0, yOffset);
                elapsedTime += Time.deltaTime;
                yield return null;
            }

            Destroy(text);
        }

        private static string GetNumber(TMP_Text text)
        {
            return text.text.Split(": ")[1];
        }

        private static void SetNumber(TMP_Text text, string type,int number)
        {
            text.text = $"{type}: {number}";
        }
    }
}
